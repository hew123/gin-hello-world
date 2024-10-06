package po

import (
	"context"

	"gorm.io/gorm"
)

type Comment struct {
	ID      uint64 `json:"id" gorm:"primaryKey"`
	Content string `json:"content"`
	PostID  uint64 `json:"post_id"`
	// TODO: add created_at to get most recent comments
}

type CommentPersistenceService struct {
	db *gorm.DB
}

func NewCommentPersistence(db *gorm.DB) CommentPersistenceService {
	return CommentPersistenceService{db: db}
}

func (c CommentPersistenceService) CreateComment(ctx context.Context, comment Comment) (Comment, error) {
	res := c.db.Create(&comment)
	if res.Error != nil {
		return comment, res.Error
	}
	return comment, nil
}
