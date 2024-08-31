package vo

import (
	"context"
	"encoding/json"
	"fmt"
	"gin-hello-world/po"
	"sync"
	"time"
)

type PostResp struct {
	Post  *po.Post
	Error error
}

var (
	postsToCreate []*po.Post
	responses     []chan PostResp
	mutex         sync.Mutex
)

type PostService struct {
}

func NewPostService(ctx context.Context, tickerDuration time.Duration) PostService {
	postService := PostService{}

	go func() {
		ticker := time.NewTicker(tickerDuration)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				postService.tickerCreate(ctx)
			}
		}
	}()

	return postService
}

type GetRankedPostsFilter struct {
	Cursor   string
	PageSize int
}

func (p PostService) GetRankedPosts(ctx context.Context, filter GetRankedPostsFilter) ([]po.Post, error) {
	rdb, err := po.GetRedisFromContext(ctx)
	if err != nil {
		return nil, err
	}
	allPosts, err := p.Find(ctx, po.FindPostFilter{})
	if err != nil {
		return nil, err
	}
	err = rdb.Set(ctx, "key1", allPosts, 0).Err()
	if err != nil {
		return nil, err
	}

	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		return nil, err
	}
	res := []po.Post{}
	err = json.Unmarshal([]byte(val), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (p PostService) Find(ctx context.Context, filter po.FindPostFilter) ([]*po.Post, error) {
	return po.FindPosts(ctx, filter)
}

func (p PostService) Create(ctx context.Context, post *po.Post, res chan PostResp) error {
	mutex.Lock()
	postsToCreate = append(postsToCreate, post)
	responses = append(responses, res)
	mutex.Unlock()
	return nil
}

func (p PostService) CreateComment(ctx context.Context, comment *po.Comment) (*po.Comment, error) {
	return po.CreateComment(ctx, comment)
}

func (p PostService) tickerCreate(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Println("posts to create: +v", postsToCreate)
	// TechDebt: this is all or nothing transaction
	// If there is one bad input, all requests will fail
	// Need to find a balance between performance and API usability
	posts, err := po.BulkCreatePosts(ctx, postsToCreate)

	// TODO: check if order is correct
	for i, post := range posts {
		// send response back to caller of Create()
		responses[i] <- PostResp{post, err}
		close(responses[i])
	}
	// This clears the slice without reallocating memory
	postsToCreate = postsToCreate[:0]
	responses = responses[:0]
}
