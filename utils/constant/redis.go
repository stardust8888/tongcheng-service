package constant

import "fmt"

const (
	//String
	RedisChatSeqID           = "ChatSeqID"
	RedisChatMaxReadSeqID    = "ChatMaxReadSeqID"
	RedisChatMaxReceiptSeqID = "ChatMaxReceiptSeqID"
	RedisWxPrepayLock        = "WxPrepayLock"
	RedisWxPayCallbackLock   = "WxPayCallbackLock"
	RedisUserViewedTimes     = "UserViewedTimes"
	RedisFeedbackDaily       = "FeedbackDaily"
	//Set
	RedisUserChatID    = "UserChatID"         //存一个用户有哪些会话
	RedisChatUserID    = "ChatUserID"         //存一个会话有哪些用户
	RedisUserBlackList = "RedisUserBlackList" //用户黑名单
	//Hset
	RedisChatLatestMessage = "ChatLatestMessage"
	//List
	RedisChatMessageQueue = "ChatMessageQueue"
)

func RedisKey(key string, args ...any) string {
	for i := range args {
		key = key + ":" + fmt.Sprint(args[i])
	}
	return key
}
