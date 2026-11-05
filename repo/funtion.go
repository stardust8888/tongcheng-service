package repo

import (
	"context"

	"go.uber.org/zap"

	"tongcheng/db/mysql"
	"tongcheng/db/redis"
	"tongcheng/library/ginx"
	"tongcheng/library/snowflake"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/utils"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

func AfterRegister(ctx context.Context, userID int) error {
	var userChats []*model.UserChat
	err := mysql.Find(ctx, mysql.GetUserDB(), &userChats, map[string]any{"user_id": userID})
	if err != nil {
		return err
	}

	chatsID := make([]any, 0)
	for i := range userChats {
		chatsID = append(chatsID, userChats[i].ChatID)
		err = redis.SAdd(ctx, constant.RedisKey(constant.RedisChatUserID, userChats[i].ChatID), userID)
		if err != nil {
			logs.Error(ctx, "FastLogin SAdd RedisChatUserID err", zap.Error(err))
		}
	}
	if len(chatsID) > 0 {
		err = redis.SAdd(ctx, constant.RedisKey(constant.RedisUserChatID, userID), chatsID...)
		if err != nil {
			logs.Error(ctx, "FastLogin SAdd RedisUserChatID err", zap.Error(err))
		}
	}

	err = LoadBlackList(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func CreateChat(ctx context.Context, userID, talkerID int) ([]*model.UserChat, error) {
	chatID := snowflake.ChatID()
	userChats := []*model.UserChat{
		{ChatID: chatID, UserID: userID, TalkerID: talkerID},
		{ChatID: chatID, UserID: talkerID, TalkerID: userID},
	}
	err := mysql.BatchCreate(ginx.Context, mysql.GetUserDB(), userChats)
	if err != nil {
		return userChats, err
	}
	_ = redis.SAdd(ctx, constant.RedisKey(constant.RedisUserChatID, userID), chatID)
	_ = redis.SAdd(ctx, constant.RedisKey(constant.RedisChatUserID, chatID), userID)
	_ = redis.SAdd(ctx, constant.RedisKey(constant.RedisUserChatID, talkerID), chatID)
	_ = redis.SAdd(ctx, constant.RedisKey(constant.RedisChatUserID, chatID), talkerID)

	return userChats, nil
}

// 把最近一百条记录刷新到缓存
func RefreshMessageCache(ctx context.Context, chatID string) error {
	messages := make([]*model.Message, 100)
	where := map[string]any{
		"chat_id":   chatID,
		mysql.Order: "seq_id desc",
	}
	mysql.SetPage(where, 1, 100)
	_, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &messages, where)
	if err != nil {
		return err
	}

	members := make([]redis.Z, 0)
	for _, msg := range messages {
		members = append(members, redis.Z{
			Score:  float64(msg.SeqID),
			Member: utils.ToJson(msg),
		})
	}
	MessageQueueKey := constant.RedisKey(constant.RedisChatMessageQueue, chatID)
	err = redis.ZAdd(context.Background(), MessageQueueKey, members...)
	if err != nil {
		logs.Error(ginx.Context, "redis.ZAdd error:", zap.Error(err))
		return err
	}
	_ = redis.Expire(context.Background(), MessageQueueKey, 3*timex.Day)

	return nil
}
