package redis

import (
	"context"

	"github.com/zjutjh/WeJH-SDK/redisHelper"
	"go.uber.org/zap"
)

// Init 初始化 Redis 连接和 Stream
// func Init() error {
func init() {
	// 初始化 Redis 客户端
	info := getConfig()
	RedisClient = redisHelper.Init(&info)

	// 测试连接
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		zap.L().Error("Failed to connect to Redis", zap.Error(err))
		panic(err)
	}

	// // 初始化 Stream 配置
	// StreamName = config.Config.GetString("redis.stream.name")
	// GroupName = config.Config.GetString("redis.stream.group")

	// // 使用 XGroupCreateMkStream 一步创建 Stream 和消费者组
	// err := RedisClient.XGroupCreateMkStream(ctx, StreamName, GroupName, "0").Err()
	// if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
	// 	zap.L().Error("Failed to create stream and consumer group", zap.Error(err))
	// 	panic(err)
	// }

	zap.L().Info("Redis initialized successfully")

	// zap.L().Info("Redis initialized successfully",
	// 	zap.String("stream", StreamName),
	// 	zap.String("group", GroupName))
}

// // PublishToStream 发布消息到 Stream
// func PublishToStream(ctx context.Context, data map[string]any) error {
// 	return RedisClient.XAdd(ctx, &redis.XAddArgs{
// 		Stream: StreamName,
// 		Values: data,
// 	}).Err()
// }

// // ConsumeFromStream 从 Stream 消费消息
// func ConsumeFromStream(ctx context.Context, consumerName string) ([]redis.XStream, error) {
// 	return RedisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
// 		Group:    GroupName,
// 		Consumer: consumerName,
// 		Streams:  []string{StreamName, ">"},
// 		Count:    1,
// 		Block:    0,
// 	}).Result()
// }

// // AckMessage 确认消息处理完成
// func AckMessage(ctx context.Context, messageID string) error {
// 	return RedisClient.XAck(ctx, StreamName, GroupName, messageID).Err()
// }
