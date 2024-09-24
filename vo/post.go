package vo

import (
	"context"
	"fmt"
	"gin-hello-world/po"
	"time"

	"github.com/redis/go-redis/v9"
)

type PostResp struct {
	Post  po.Post
	Error error
}

type PostService struct {
}

func NewPostService(ctx context.Context, tickerDuration time.Duration) PostService {
	postService := PostService{}
	// TechDebt: this should be pushed to a daemon instance (singleton)
	// separated from server instance which is to be scaled
	postService.BulkInitRankedPosts(ctx)

	go func() {
		ticker := time.NewTicker(tickerDuration)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				err := postService.BulkResetRankedPosts(ctx)
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	return postService
}

type GetRankedPostsFilter = po.GetRankedPostsFilter
type FindPostFilter = po.FindPostFilter

type GetRankedPostsResp struct {
	Posts   []po.PostWithScore `json:"posts"`
	Version int64              `json:"version"`
	Cursor  int                `json:"cursor"`
	Count   int                `json:"count"`
}

func (p PostService) GetRankedPosts(ctx context.Context, filter GetRankedPostsFilter) (*GetRankedPostsResp, error) {
	// User may not have the latest version. Fetch it for them
	if filter.Version == 0 {
		ver, err := po.GetLatestSnapshotVersion(ctx)
		if err != nil && err != redis.Nil {
			return nil, err
		}
		filter.Version = ver
	}
	posts, err := po.GetRankedPosts(ctx, filter)
	if err != nil {
		return nil, err
	}
	count := len(posts)
	return &GetRankedPostsResp{
		Posts:   posts,
		Count:   count,
		Version: filter.Version,
		Cursor:  filter.Cursor + count,
	}, nil
}

func (p PostService) Find(ctx context.Context, filter FindPostFilter) ([]Post, error) {
	return po.FindPosts(ctx, filter)
}
