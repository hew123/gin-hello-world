package po

import (
	"context"

	"gorm.io/gorm"
)

type Post struct {
	ID      uint64 `json:"id" gorm:"primaryKey"`
	Caption string `json:"caption"`
	// TODO: check why foreign key constraint not working
	Comments []Comment `json:"comments" gorm:"foreignKey:PostID;references:ID"`
}

type PostPersistenceService struct {
	db *gorm.DB
}

type FindPostFilter struct {
	PostIDs *[]uint64
}

func NewPostPersistence(db *gorm.DB) PostPersistenceService {
	return PostPersistenceService{db: db}
}

func (p PostPersistenceService) CreatePost(ctx context.Context, post Post) (Post, error) {
	res := p.db.Create(&post)
	if res.Error != nil {
		return post, res.Error
	}
	return post, nil
}

func (p PostPersistenceService) BulkCreatePosts(ctx context.Context, posts []Post) ([]Post, error) {
	tx := p.db.Begin()
	if tx.Error != nil {
		return posts, tx.Error
	}
	res := tx.Create(&posts)
	if res.Error != nil {
		tx.Rollback()
		return posts, res.Error
	}
	return posts, tx.Commit().Error
}

func (p PostPersistenceService) FindPosts(ctx context.Context, filter FindPostFilter) ([]Post, error) {
	db := p.db.Preload("Comments")
	posts := []Post{}
	if filter.PostIDs != nil {
		db = db.Where("id IN ?", *filter.PostIDs)
	}
	res := db.Find(&posts)
	if res.Error != nil {
		return nil, res.Error
	}
	return posts, nil
}
