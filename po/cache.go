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

func IncSnapshotVersion(ctx context.Context) (int64, error) {
	rdb, err := GetRedisFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return rdb.Incr(ctx, PostSnapshotVer).Result()
}

func GetLatestSnapshotVersion(ctx context.Context) (int64, error) {
	rdb, err := GetRedisFromContext(ctx)
	if err != nil {
		return 0, err
	}
	val, err := rdb.Get(ctx, PostSnapshotVer).Result()
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, err
}

func SetTTL(ctx context.Context, version int64) error {
	rdb, err := GetRedisFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = rdb.Expire(ctx, fmt.Sprintf(RankedPostsKey, version), TTL).Result()
	return err
}

func CopyToNewSnapshot(ctx context.Context, sourceVer int64, targetVer int64) error {
	rdb, err := GetRedisFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = rdb.Copy(ctx,
		fmt.Sprintf(RankedPostsKey, sourceVer),
		fmt.Sprintf(RankedPostsKey, targetVer),
		0, false).Result()
	return err
}

func BulkSetRankedPosts(ctx context.Context, version int64, posts *[]PostWithScore) error {
	rdb, err := GetRedisFromContext(ctx)
	if err != nil {
		return err
	}
	for _, post := range *posts {
		jsonStr, err := json.Marshal(post)
		if err != nil {
			return err
		}
		// TODO: use pipeline to do batch insert
		err = rdb.ZAdd(
			ctx, fmt.Sprintf(RankedPostsKey, version),
			redis.Z{
				Score:  float64(post.Score),
				Member: jsonStr,
			}).Err()
		if err != nil {
			return err
		}
	}
	return nil
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

func GetRankedPosts(ctx context.Context, filter GetRankedPostsFilter) ([]PostWithScore, error) {
	rdb, err := GetRedisFromContext(ctx)
	if err != nil {
		return nil, err
	}
	valuesWithScore, err := rdb.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
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
