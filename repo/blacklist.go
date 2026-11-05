package repo

import (
	"context"

	"tongcheng/db/mysql"
	"tongcheng/db/redis"
	"tongcheng/model"
	"tongcheng/utils/constant"
)

// LoadBlackList 从DB加载黑名单到Redis
func LoadBlackList(ctx context.Context, userID int) error {
	records := make([]*model.UserBlackList, 0)
	err := mysql.Find(ctx, mysql.GetUserDB(), &records, map[string]any{"user_id": userID})
	if err != nil {
		return err
	}
	key := constant.RedisKey(constant.RedisUserBlackList, userID)
	if len(records) == 0 {
		return nil
	}
	members := make([]any, 0, len(records))
	for _, r := range records {
		members = append(members, r.BlackUserID)
	}
	return redis.SAdd(ctx, key, members...)
}

// AddBlackCache 加入黑名单缓存
func AddBlackCache(ctx context.Context, userID, blackUserID int) error {
	key := constant.RedisKey(constant.RedisUserBlackList, userID)
	return redis.SAdd(ctx, key, blackUserID)
}

// RemoveBlackCache 移除黑名单缓存
func RemoveBlackCache(ctx context.Context, userID, blackUserID int) error {
	key := constant.RedisKey(constant.RedisUserBlackList, userID)
	return redis.SRem(ctx, key, blackUserID)
}

// IsBlacked 检查userID是否把blackID拉黑了
func IsBlacked(ctx context.Context, userID, blackID int) (bool, error) {
	key := constant.RedisKey(constant.RedisUserBlackList, userID)
	return redis.SIsMember(ctx, key, blackID)
}
