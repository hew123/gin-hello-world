package po

import (
	"context"
)

type Comment struct {
	ID      uint64 `json:"id" gorm:"primaryKey"`
	Content string `json:"content"`
	PostID  uint64 `json:"post_id"`
}

func CreateComment(ctx context.Context, comment Comment) (Comment, error) {
	db, err := GetDbFromContext(ctx)
	if err != nil {
		return comment, err
	}
	res := db.Create(&comment)
	if res.Error != nil {
		return comment, res.Error
	}
	return comment, nil
}
