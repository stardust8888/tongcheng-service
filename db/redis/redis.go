package redis

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"

	"tongcheng/conf"
)

var (
	client      redis.UniversalClient
	Nil         = redis.Nil
	ForeverTime = time.Duration(0)
)

func Init() error {
	config := conf.GetConfiguration().Redis
	client = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:           config.Addresses,
		Username:        config.Username,
		Password:        config.Password,
		PoolSize:        50,
		MinIdleConns:    10,
		MaxIdleConns:    30,
		ConnMaxIdleTime: 5 * time.Minute,  // 空闲连接超时回收
		ConnMaxLifetime: 30 * time.Minute, // 防止连接老化
		DialTimeout:     3 * time.Second,
		ReadTimeout:     500 * time.Millisecond,
		WriteTimeout:    500 * time.Millisecond,
		PoolTimeout:     1 * time.Second, // 等待空闲连接的超时
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis连接失败: %v", err)
	}

	return nil
}

func GetRedisClient() redis.UniversalClient {
	return client
}

// Close 关闭Redis客户端连接
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// Set 设置键值对
func Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return client.Set(ctx, key, value, expiration).Err()
}

func SetNX(ctx context.Context, key string, value any, expiration time.Duration) error {
	return client.SetNX(ctx, key, value, expiration).Err()
}

// Get 获取键对应的值
func Get(ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

// Del 删除一个或多个键
func Del(ctx context.Context, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	result, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Expire 设置键的过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return client.Expire(ctx, key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func TTL(ctx context.Context, key string) (time.Duration, error) {
	return client.TTL(ctx, key).Result()
}

// LPush 将一个或多个值插入到列表头部
func LPush(ctx context.Context, key string, values ...any) error {
	return client.LPush(ctx, key, values...).Err()
}

// RPush 将一个或多个值插入到列表尾部
func RPush(ctx context.Context, key string, values ...any) error {
	return client.RPush(ctx, key, values...).Err()
}

// LPop 移除并返回列表的第一个元素
func LPop(ctx context.Context, key string) (string, error) {
	return client.LPop(ctx, key).Result()
}

// RPop 移除并返回列表的最后一个元素
func RPop(ctx context.Context, key string) (string, error) {
	return client.RPop(ctx, key).Result()
}

// LRange 获取列表指定范围内的元素
func LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return client.LRange(ctx, key, start, stop).Result()
}

// LLen 获取列表长度
func LLen(ctx context.Context, key string) (int64, error) {
	return client.LLen(ctx, key).Result()
}

// LRem 根据参数 COUNT 的值，移除列表中与参数 VALUE 相等的元素
func LRem(ctx context.Context, key string, count int64, value any) error {
	return client.LRem(ctx, key, count, value).Err()
}

// LIndex 通过索引获取列表中的元素
func LIndex(ctx context.Context, key string, index int64) (string, error) {
	return client.LIndex(ctx, key, index).Result()
}

// LSet 通过索引设置列表元素的值
func LSet(ctx context.Context, key string, index int64, value any) error {
	return client.LSet(ctx, key, index, value).Err()
}

// LTrim 对一个列表进行修剪(trim)，让列表只保留指定区间内的元素
func LTrim(ctx context.Context, key string, start, stop int64) error {
	return client.LTrim(ctx, key, start, stop).Err()
}

// LInsert 将值插入到列表中指定元素的前面或后面
func LInsert(ctx context.Context, key string, op string, pivot, value any) error {
	return client.LInsert(ctx, key, op, pivot, value).Err()
}

// HSet 设置哈希表字段的值
func HSet(ctx context.Context, key string, field string, value any) error {
	return client.HSet(ctx, key, field, value).Err()
}

// HGet 获取哈希表字段的值
func HGet(ctx context.Context, key, field string) (string, error) {
	return client.HGet(ctx, key, field).Result()
}

// HMSet 设置多个哈希表字段的值
func HMSet(ctx context.Context, key string, fields map[string]any) error {
	return client.HMSet(ctx, key, fields).Err()
}

// HMGet 获取多个哈希表字段的值
func HMGet(ctx context.Context, key string, fields ...string) ([]any, error) {
	return client.HMGet(ctx, key, fields...).Result()
}

// HDel 删除哈希表一个或多个字段
func HDel(ctx context.Context, key string, fields ...string) error {
	return client.HDel(ctx, key, fields...).Err()
}

// HExists 检查哈希表字段是否存在
func HExists(ctx context.Context, key, field string) (bool, error) {
	return client.HExists(ctx, key, field).Result()
}

// HGetAll 获取哈希表所有字段和值
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return client.HGetAll(ctx, key).Result()
}

// SAdd 向集合添加一个或多个成员
func SAdd(ctx context.Context, key string, members ...any) error {
	return client.SAdd(ctx, key, members...).Err()
}

// SRem 移除集合中一个或多个成员
func SRem(ctx context.Context, key string, members ...any) error {
	return client.SRem(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func SMembers(ctx context.Context, key string) ([]string, error) {
	return client.SMembers(ctx, key).Result()
}

// SIsMember 判断成员是否是集合的成员
func SIsMember(ctx context.Context, key string, member any) (bool, error) {
	return client.SIsMember(ctx, key, member).Result()
}

type Z struct {
	Score  float64
	Member any
}

// ZAdd 向有序集合添加一个或多个成员
func ZAdd(ctx context.Context, key string, members ...Z) error {
	zs := make([]redis.Z, len(members))
	for i := range members {
		zs[i].Score = members[i].Score
		zs[i].Member = members[i].Member
	}
	return client.ZAdd(ctx, key, zs...).Err()
}

// ZCard 获取有序集合的成员数
func ZCard(ctx context.Context, key string) (int64, error) {
	return client.ZCard(ctx, key).Result()
}

// ZRange 获取有序集合的指定范围成员
func ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return client.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange 获取有序集合的指定范围成员(按分数从高到低)
func ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return client.ZRevRange(ctx, key, start, stop).Result()
}

// ZMaxScoreMember 获取有序集合中分数最大的成员
func ZMaxScoreMember(ctx context.Context, key string) (string, error) {
	members, err := client.ZRevRange(ctx, key, 0, 0).Result()
	if err != nil {
		return "", err
	}
	if len(members) == 0 {
		return "", redis.Nil
	}
	return members[0], nil
}

// ZRem 移除有序集合中的一个或多个成员
func ZRem(ctx context.Context, key string, members ...any) error {
	return client.ZRem(ctx, key, members...).Err()
}

// ZRemRangeByRank 移除有序集合中给定排名区间的所有成员
func ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	return client.ZRemRangeByRank(ctx, key, start, stop).Result()
}

// ZTrim 保留有序集合中分数最高的count个元素，移除其他元素
func ZTrim(ctx context.Context, key string, count int64) (int64, error) {
	// 使用ZCard获取有序集合的总数量
	total, err := client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// 如果总数量小于等于要保留的数量，则不需要裁剪
	if total <= count {
		return 0, nil
	}

	// 计算需要移除的元素数量
	removeCount := total - count

	// 移除分数最低的元素（排名从0到removeCount-1）
	return client.ZRemRangeByRank(ctx, key, 0, removeCount-1).Result()
}

// Incr 将键的值增加1
func Incr(ctx context.Context, key string) (int64, error) {
	return client.Incr(ctx, key).Result()
}

// Decr 将键的值减少1
func Decr(ctx context.Context, key string) (int64, error) {
	return client.Decr(ctx, key).Result()
}

// Ping 测试Redis连接
func Ping(ctx context.Context) error {
	return client.Ping(ctx).Err()
}

// XAdd 向Stream添加消息
func XAdd(ctx context.Context, stream string, values map[string]any) (string, error) {
	return client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Result()
}

// XRange 获取Stream中指定范围内的消息
func XRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error) {
	return client.XRange(ctx, stream, start, stop).Result()
}

// XRevRange 获取Stream中指定范围内的消息(反向)
func XRevRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error) {
	return client.XRevRange(ctx, stream, start, stop).Result()
}

// XRead 从一个或多个Stream中读取消息
func XRead(ctx context.Context, streams ...string) ([]redis.XStream, error) {
	return client.XRead(ctx, &redis.XReadArgs{
		Streams: streams,
	}).Result()
}

// XReadGroup 从Stream中消费消息(消费者组模式)
func XReadGroup(ctx context.Context, group, consumer string, streams ...string) ([]redis.XStream, error) {
	return client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  streams,
	}).Result()
}

// XGroupCreate 创建消费者组
func XGroupCreate(ctx context.Context, stream, group, start string) error {
	return client.XGroupCreate(ctx, stream, group, start).Err()
}

// XGroupCreateMkStream 创建消费者组，如果Stream不存在则创建
func XGroupCreateMkStream(ctx context.Context, stream, group, start string) error {
	return client.XGroupCreateMkStream(ctx, stream, group, start).Err()
}

// XGroupDestroy 销毁消费者组
func XGroupDestroy(ctx context.Context, stream, group string) (int64, error) {
	return client.XGroupDestroy(ctx, stream, group).Result()
}

// XAck 确认消息处理完成
func XAck(ctx context.Context, stream, group string, ids ...string) error {
	return client.XAck(ctx, stream, group, ids...).Err()
}

// XPending 获取待处理消息的信息
func XPending(ctx context.Context, stream, group string) (*redis.XPending, error) {
	return client.XPending(ctx, stream, group).Result()
}

// XPendingExt 获取待处理消息的详细信息
func XPendingExt(ctx context.Context, stream, group, start, end string, count int64, consumer string) ([]redis.XPendingExt, error) {
	return client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream:   stream,
		Group:    group,
		Start:    start,
		End:      end,
		Count:    count,
		Consumer: consumer,
	}).Result()
}

// XClaim 转移消息的归属权
func XClaim(ctx context.Context, stream, group, consumer string, minIdle time.Duration, ids ...string) ([]redis.XMessage, error) {
	return client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   stream,
		Group:    group,
		Consumer: consumer,
		MinIdle:  minIdle,
		Messages: ids,
	}).Result()
}

// XDel 从Stream中删除消息
func XDel(ctx context.Context, stream string, ids ...string) error {
	return client.XDel(ctx, stream, ids...).Err()
}

// XLen 获取Stream的长度
func XLen(ctx context.Context, stream string) (int64, error) {
	return client.XLen(ctx, stream).Result()
}
