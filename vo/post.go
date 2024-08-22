package vo

import (
	"fmt"
	"gin-hello-world/po"
	"sync"
	"time"
)

var (
	postsToCreate []*po.Post
	responses     []chan *po.Post
)

type PostService struct {
	persistenceService po.PostPersistenceService
	//postsToCreate      []*po.Post
	mutex sync.Mutex
	//responses []chan *po.Post
}

func NewPostService(dbName string, tickerDuration time.Duration) PostService {
	postService := PostService{
		//postsToCreate:      []*po.Post{},
		persistenceService: po.NewPostPersistenceService(dbName),
		mutex:              sync.Mutex{},
		//responses:          []chan *po.Post{},
	}

	go func() {
		ticker := time.NewTicker(tickerDuration)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				err := postService.tickerCreate()
				if err != nil {
					fmt.Println("error creating posts: ", err)
				} else {
					fmt.Println("ticker-created posts")
				}
			}
		}
	}()

	return postService
}

func (p PostService) Find(filter po.FindPostFilter) ([]*po.Post, error) {
	return p.persistenceService.Find(filter)
}

func (p PostService) Create(post *po.Post, res chan *po.Post) error {
	p.mutex.Lock()
	fmt.Println("Appending post", *post)
	postsToCreate = append(postsToCreate, post)
	fmt.Println("Appended post", postsToCreate)
	responses = append(responses, res)
	p.mutex.Unlock()
	return nil
}

func (p PostService) tickerCreate() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	fmt.Printf("posts to create: +v", postsToCreate)
	posts, err := p.persistenceService.BulkCreate(postsToCreate)
	fmt.Printf("posts created: +v", posts)
	if err != nil {
		return err
	}
	// TODO: check if order is correct
	for i, post := range posts {
		// send response back to caller of Create()
		responses[i] <- post
		close(responses[i])
	}
	// This clears the slice without reallocating memory
	postsToCreate = postsToCreate[:0]
	responses = responses[:0]
	return nil
}
