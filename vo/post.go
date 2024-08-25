package vo

import (
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
	persistenceService po.PostPersistenceService
}

func NewPostService(dbName string, tickerDuration time.Duration) PostService {
	postService := PostService{
		persistenceService: po.NewPostPersistenceService(dbName),
	}

	go func() {
		ticker := time.NewTicker(tickerDuration)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				postService.tickerCreate()
			}
		}
	}()

	return postService
}

func (p PostService) Find(filter po.FindPostFilter) ([]*po.Post, error) {
	return p.persistenceService.Find(filter)
}

func (p PostService) Create(post *po.Post, res chan PostResp) error {
	mutex.Lock()
	postsToCreate = append(postsToCreate, post)
	responses = append(responses, res)
	mutex.Unlock()
	return nil
}

func (p PostService) tickerCreate() {
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Println("posts to create: +v", postsToCreate)
	posts, err := p.persistenceService.BulkCreate(postsToCreate)

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
