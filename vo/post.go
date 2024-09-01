package vo

import (
	"context"
	"encoding/json"
	"fmt"
	"gin-hello-world/po"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
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

func (p PostService) GetRankedPosts(ctx context.Context, filter GetRankedPostsFilter) ([]*po.Post, error) {
	rdb, err := po.GetRedisFromContext(ctx)
	if err != nil {
		return nil, err
	}
	allPosts, err := p.Find(ctx, po.FindPostFilter{})
	if err != nil {
		return nil, err
	}
	for _, post := range allPosts {
		jsonStr, err := json.Marshal(post)
		if err != nil {
			return nil, err
		}
		err = rdb.ZAdd(ctx, "key1", redis.Z{Score: float64(len(post.Comments)), Member: jsonStr}).Err()
		if err != nil {
			return nil, err
		}
	}

	valuesWithScore, err := rdb.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
		Key:   "key1",
		Start: 0,
		Stop:  10,
		//ByScore: true,
		//Offset:  10,
		//Count:   0,
		Rev: true,
	}).Result()

	if err != nil {
		return nil, err
	}
	res := []*po.Post{}
	for _, valueWithScore := range valuesWithScore {
		post := po.Post{}
		err = json.Unmarshal([]byte(valueWithScore.Member.(string)), &post)
		if err != nil {
			return nil, err
		}
		res = append(res, &post)
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
