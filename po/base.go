package po

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

const (
	DbContextKey = "database"
)

func SetDbInContext(c context.Context, db *gorm.DB) context.Context {
	return context.WithValue(c, DbContextKey, db)
}

func GetDbFromContext(c context.Context) (*gorm.DB, error) {
	val := c.Value(DbContextKey).(*gorm.DB)
	if val == nil {
		return nil, errors.New("DB context not set")
	}
	return val, nil
}
