package po

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Post struct {
	ID      uint64 `json:"id"`
	Caption string `json:"caption"`
}

type PostPersistenceService struct {
	db *gorm.DB
}

func NewPostPersistenceService(dbName string) PostPersistenceService {
	dbConn, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return PostPersistenceService{
		db: dbConn,
	}
}

type FindPostFilter struct {
	PostIDs *[]uint64
}

func (s PostPersistenceService) Create(post *Post) (*Post, error) {
	res := s.db.Create(post)
	if res.Error != nil {
		return nil, res.Error
	}
	return post, nil
}

func (s PostPersistenceService) BulkCreate(posts []*Post) ([]*Post, error) {
	tx := s.db.Begin()
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

func (s PostPersistenceService) Find(filter FindPostFilter) ([]*Post, error) {
	posts := []*Post{}
	db := s.db
	if filter.PostIDs != nil {
		db = db.Where("id IN ?", *filter.PostIDs)
	}
	res := db.Find(&posts)
	if res.Error != nil {
		return nil, res.Error
	}
	return posts, nil
}
