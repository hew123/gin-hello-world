package po

import (
	"context"
)

type Post struct {
	ID      uint64 `json:"id" gorm:"primaryKey"`
	Caption string `json:"caption"`
	// TODO: check why foreign key constraint not working
	Comments []Comment `json:"comments" gorm:"foreignKey:PostID;references:ID"`
}

type FindPostFilter struct {
	PostIDs *[]uint64
}

func Create(ctx context.Context, post Post) (Post, error) {
	db, err := GetDbFromContext(ctx)
	if err != nil {
		return post, err
	}
	res := db.Create(post)
	if res.Error != nil {
		return post, res.Error
	}
	return post, nil
}

func BulkCreatePosts(ctx context.Context, posts []Post) ([]Post, error) {
	db, err := GetDbFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tx := db.Begin()
	if tx.Error != nil {
		return posts, tx.Error
	}
	res := tx.Create(posts)
	if res.Error != nil {
		tx.Rollback()
		return posts, res.Error
	}
	return posts, tx.Commit().Error
}

func FindPosts(ctx context.Context, filter FindPostFilter) ([]Post, error) {
	db, err := GetDbFromContext(ctx)
	if err != nil {
		return nil, err
	}
	db = db.Preload("Comments")
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
