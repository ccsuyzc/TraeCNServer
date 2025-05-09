package redis_service

import (
	. "TraeCNServer/db"
	"TraeCNServer/model"
	"context"
	"strconv"

	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	articleLikeHash   = "article_likes"
	userLikeSetPrefix = "user_likes:"
	syncBatchSize     = 100
)

// 点赞文章
func LikeArticle(ctx context.Context, articleID uint, userID uint) error {
	// 使用事务确保原子性
	userKey := userLikeSetPrefix + strconv.FormatUint(uint64(articleID), 10)

	// 使用事务保证原子性
	_, err := RedisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 检查是否已点赞
		if pipe.SIsMember(ctx, userKey, userID).Val() {
			return errors.New("用户已点赞过该文章")
		}
		// 记录用户点赞
		pipe.SAdd(ctx, userKey, userID)
		// 增加文章点赞计数
		pipe.HIncrBy(ctx, articleLikeHash, strconv.FormatUint(uint64(articleID), 10), 1)
		return nil
	})
	return err
}

// 取消点赞
func UnlikeArticle(ctx context.Context, articleID uint, userID uint) error {
	userKey := userLikeSetPrefix + strconv.FormatUint(uint64(articleID), 10)

	_, err := RedisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 移除用户点赞记录
		if pipe.SRem(ctx, userKey, userID).Val() == 0 {
			return errors.New("用户未点赞过该文章")
		}
		// 减少文章点赞计数
		pipe.HIncrBy(ctx, articleLikeHash, strconv.FormatUint(uint64(articleID), 10), -1)
		return nil
	})
	return err
}

// 获取文章点赞数
func GetArticleLikes(ctx context.Context, articleID uint) (int64, error) {
	countStr, err := RedisClient.HGet(ctx, articleLikeHash, strconv.FormatUint(uint64(articleID), 10)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	return strconv.ParseInt(countStr, 10, 64)
}

// 同步到数据库
func SyncLikesToDB(ctx context.Context) error {
	// 分批处理逻辑
	cursor := uint64(0)
	for {
		// 分批扫描哈希表
		fields, cur, err := RedisClient.HScan(ctx, articleLikeHash, cursor, "*", syncBatchSize).Result()
		if err != nil {
			return err
		}

		// 处理批量数据
		for i := 0; i < len(fields); i += 2 {
			articleID, _ := strconv.ParseUint(fields[i], 10, 64)
			likeCount, _ := strconv.ParseInt(fields[i+1], 10, 64)

			// 更新数据库逻辑
			DB.Model(&model.Article{}).
				Where("id = ?", uint(articleID)).
				UpdateColumn("like_count", gorm.Expr("like_count + ?", likeCount))
		}

		if cur == 0 {
			break
		}
		cursor = cur
	}
	return nil
}
