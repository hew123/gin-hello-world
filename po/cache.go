package po

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// e.g. RANKED_POSTS:v1
	RankedPostsKey  = "RANKED_POSTS:%v"
	PostSnapshotVer = "POST_SNAPSHOT_VERSION"
	MaxPageNumber   = 100
	TTL             = 10 * time.Minute
)

type CachingService struct {
	rdb *redis.Client
}

func NewCachingService(rdb *redis.Client) CachingService {
	return CachingService{rdb: rdb}
}

func (c CachingService) IncSnapshotVersion(ctx context.Context) (int64, error) {
	return c.rdb.Incr(ctx, PostSnapshotVer).Result()
}

func (c CachingService) GetLatestSnapshotVersion(ctx context.Context) (int64, error) {
	val, err := c.rdb.Get(ctx, PostSnapshotVer).Result()
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, err
}

func (c CachingService) SetTTL(ctx context.Context, version int64) error {
	_, err := c.rdb.Expire(ctx, fmt.Sprintf(RankedPostsKey, version), TTL).Result()
	return err
}

func (c CachingService) CopyToNewSnapshot(ctx context.Context, sourceVer int64, targetVer int64) error {
	_, err := c.rdb.Copy(ctx,
		fmt.Sprintf(RankedPostsKey, sourceVer),
		fmt.Sprintf(RankedPostsKey, targetVer),
		0, false).Result()
	return err
}

func (c CachingService) BulkSetRankedPosts(ctx context.Context, version int64, posts *[]PostWithScore) error {
	redisVals := []redis.Z{}
	if posts == nil || len(*posts) == 0 {
		return nil
	}
	for _, post := range *posts {
		jsonStr, err := json.Marshal(post)
		if err != nil {
			return err
		}
		redisVals = append(redisVals, redis.Z{
			Score:  float64(post.Score),
			Member: jsonStr,
		})
	}
	// TODO: Replace exising post by ID instead of creating new element
	// most recent 2 comments change - so cannot rely on post being same json string
	return c.rdb.ZAdd(
		ctx, fmt.Sprintf(RankedPostsKey, version),
		redisVals...,
	).Err()
}

type GetRankedPostsFilter struct {
	Version int64 `form:"version"`
	Cursor  int   `form:"cursor"`
	Count   int   `form:"count" binding:"required"`
}

type PostWithScore struct {
	Post
	Score int64 `json:"score"`
}

func (c CachingService) GetRankedPosts(ctx context.Context, filter GetRankedPostsFilter) ([]PostWithScore, error) {
	valuesWithScore, err := c.rdb.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
		Key:   fmt.Sprintf(RankedPostsKey, filter.Version),
		Start: filter.Cursor,
		Stop:  filter.Cursor + filter.Count,
		// ByScore: true,
		// Offset:  10,
		// ByLex: true,
		// Count: int64(filter.Count),
		Rev: true,
	}).Result()

	if err != nil {
		return nil, err
	}
	res := []PostWithScore{}
	for _, valueWithScore := range valuesWithScore {
		post := Post{}
		err = json.Unmarshal([]byte(valueWithScore.Member.(string)), &post)
		if err != nil {
			return nil, err
		}
		res = append(res, PostWithScore{Post: post, Score: int64(valueWithScore.Score)})
	}
	return res, nil
}
