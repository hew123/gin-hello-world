package po

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	DbContextKey      = "database"
	RedisDbContextKey = "redis"
)

func SetDbInContext(c context.Context, db *gorm.DB) context.Context {
	return context.WithValue(c, DbContextKey, db)
}

func GetDbFromContext(c context.Context) (*gorm.DB, error) {
	val := c.Value(DbContextKey).(*gorm.DB)
	if val == nil {
		return nil, errors.New("db context not set")
	}
	return val, nil
}

func SetRedisInContext(c context.Context, rdb *redis.Client) context.Context {
	return context.WithValue(c, RedisDbContextKey, rdb)
}

func GetRedisFromContext(c context.Context) (*redis.Client, error) {
	val := c.Value(RedisDbContextKey).(*redis.Client)
	if val == nil {
		return nil, errors.New("redis context not set")
	}
	return val, nil
}
