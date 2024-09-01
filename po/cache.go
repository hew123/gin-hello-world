package po

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	// e.g. RANKED_POSTS:v1
	RankedPostsKey  = "RANKED_POSTS:%v"
	PostSnapshotVer = "POST_SNAPSHOT_VERSION"
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

func BulkSetRankedPosts(ctx context.Context, version int64, posts []*Post) error {
	rdb, err := GetRedisFromContext(ctx)
	if err != nil {
		return err
	}
	for _, post := range posts {
		jsonStr, err := json.Marshal(post)
		if err != nil {
			return err
		}
		err = rdb.ZAdd(
			ctx, fmt.Sprintf(RankedPostsKey, version),
			redis.Z{
				Score:  float64(len(post.Comments)),
				Member: jsonStr,
			}).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

type GetRankedPostsFilter struct {
	Version int64
	Start   int
	Stop    int
}

func GetRankedPosts(ctx context.Context, filter GetRankedPostsFilter) ([]*Post, error) {
	rdb, err := GetRedisFromContext(ctx)
	if err != nil {
		return nil, err
	}
	valuesWithScore, err := rdb.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
		Key:   fmt.Sprintf(RankedPostsKey, filter.Version),
		Start: filter.Start,
		Stop:  filter.Stop,
		//ByScore: true,
		//Offset:  10,
		//Count:   0,
		Rev: true,
	}).Result()

	if err != nil {
		return nil, err
	}
	res := []*Post{}
	for _, valueWithScore := range valuesWithScore {
		post := Post{}
		err = json.Unmarshal([]byte(valueWithScore.Member.(string)), &post)
		if err != nil {
			return nil, err
		}
		res = append(res, &post)
	}
	return res, nil
}
