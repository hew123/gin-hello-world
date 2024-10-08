package vo

import (
	"context"
	"fmt"
	"gin-hello-world/po"
	"sync"
)

const NumOfMostRecentCommentPerPost = 2

type Post = po.Post
type Comment = po.Comment
type PostWithScore = po.PostWithScore

var (
	// TechDebt: should be pushed to a kafka instance instead
	// separated from server instance
	createdPostIDs []uint64
	mutex          sync.Mutex
)

func (p PostService) BulkInitRankedPosts(ctx context.Context) error {
	posts, err := p.postPo.FindPosts(ctx, po.FindPostFilter{})
	if err != nil {
		return err
	}
	snapShotVer, err := p.cachingPo.IncSnapshotVersion(ctx)
	if err != nil {
		return err
	}
	postsWithScore := []PostWithScore{}
	for _, p := range posts {
		commentLen := len(p.Comments)
		// Requirement: only show the most recent 2 comments
		if commentLen > NumOfMostRecentCommentPerPost {
			p.Comments = p.Comments[:NumOfMostRecentCommentPerPost]
		}
		postsWithScore = append(postsWithScore, po.PostWithScore{
			Post:  p,
			Score: int64(commentLen),
		})
	}
	fmt.Println("Latest Post snapshot: ", snapShotVer)
	err = p.cachingPo.BulkSetRankedPosts(ctx, snapShotVer, &postsWithScore)
	if err != nil {
		return err
	}
	return p.cachingPo.SetTTL(ctx, snapShotVer)
}

func (p PostService) BulkResetRankedPosts(ctx context.Context) error {
	currSnapShotVer, err := p.cachingPo.GetLatestSnapshotVersion(ctx)
	if err != nil {
		return err
	}
	newSnapShotVer, err := p.cachingPo.IncSnapshotVersion(ctx)
	if err != nil {
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	fmt.Println("Latest Post snapshot: ", newSnapShotVer)
	err = p.cachingPo.CopyToNewSnapshot(ctx, currSnapShotVer, newSnapShotVer)
	if err != nil {
		return nil
	}

	// Fetching recent posts again because their comment counts might have been updated

	posts, err := p.postPo.FindPosts(ctx, FindPostFilter{PostIDs: &createdPostIDs})
	fmt.Printf("Recent posts: %+v", createdPostIDs)
	if err != nil {
		return err
	}
	postsWithScore := []PostWithScore{}
	for _, p := range posts {
		commentLen := len(p.Comments)
		// Requirement: only show the most recent 2 comments
		if commentLen > NumOfMostRecentCommentPerPost {
			p.Comments = p.Comments[:NumOfMostRecentCommentPerPost]
		}
		postsWithScore = append(postsWithScore, PostWithScore{
			Post:  p,
			Score: int64(commentLen),
		})
	}
	err = p.cachingPo.BulkSetRankedPosts(ctx, newSnapShotVer, &postsWithScore)
	if err != nil {
		return err
	}
	createdPostIDs = make([]uint64, 0)
	return p.cachingPo.SetTTL(ctx, newSnapShotVer)
}

func (p PostService) CreatePost(ctx context.Context, post Post) (Post, error) {
	resp, err := p.postPo.CreatePost(ctx, post)
	if err != nil {
		return post, err
	}
	mutex.Lock()
	defer mutex.Unlock()
	// TODO: emit kakfa event
	createdPostIDs = append(createdPostIDs, post.ID)
	return resp, nil
}

func (p PostService) CreateComment(ctx context.Context, comment Comment) (Comment, error) {
	resp, err := p.commentPo.CreateComment(ctx, comment)
	if err != nil {
		return comment, err
	}
	mutex.Lock()
	defer mutex.Unlock()
	// TODO: emit kakfa event
	createdPostIDs = append(createdPostIDs, comment.PostID)
	return resp, nil
}
