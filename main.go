package main

import (
	"context"
	"gin-hello-world/handler"
	"gin-hello-world/po"
	"gin-hello-world/vo"
	"time"

	"github.com/gin-gonic/gin"
)

// config
const (
	DbName             = "test.db"
	PostCreateInterval = time.Second * 10
)

func main() {
	router := gin.Default()
	db := po.InitDb(DbName)
	rdb := po.InitRedis()
	ctx := context.Background()
	postService := vo.NewPostService(
		ctx,
		PostCreateInterval,
		po.NewPostPersistence(db),
		po.NewCommentPersistence(db),
		po.NewCachingService(rdb),
	)
	handler := handler.NewHandler(postService)

	router.GET("/post/get", handler.GetPosts)
	router.GET("/post/get_ranked_posts", handler.GetRankedPosts)
	router.POST("/post/create", handler.CreatePost)
	router.POST("/comment/create", handler.CreateComment)

	router.Run()
}
