package main

import (
	"context"
	"gin-hello-world/po"
	"gin-hello-world/vo"
	"net/http"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// config
const (
	DbName             = "test.db"
	PostCreateInterval = time.Second * 10
)

// Global DB instance
var db *gorm.DB
var redisDb *redis.Client
var err error

func main() {
	router := gin.Default()

	db, err = gorm.Open(sqlite.Open(DbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	s, err := miniredis.Run()
	if err != nil {
		panic("failed to init redis instance")
	}
	redisDb = redis.NewClient(&redis.Options{
		Addr:     s.Addr(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	ctx := po.SetDbInContext(context.Background(), db)
	redisCtx := po.SetRedisInContext(ctx, redisDb)
	postService := vo.NewPostService(redisCtx, PostCreateInterval)
	handler := Handler{postService}

	// TODO: add auth middleware
	router.GET("/post/get", handler.GetPosts)
	router.GET("/post/get_ranked_posts", handler.GetRankedPosts)
	router.POST("/post/create", handler.CreatePost)
	router.POST("/comment/create", handler.CreateComment)

	router.Run()
}

type Handler struct {
	PostService vo.PostService
}

type GetPostsReq struct {
	PostIDs *[]uint64 `form:"post_ids, omitempty"`
}

func (h Handler) GetPosts(c *gin.Context) {
	req := GetPostsReq{}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	// TODO: pull out set context to middleware
	ctx := po.SetDbInContext(c.Request.Context(), db)
	posts, err := h.PostService.Find(ctx, po.FindPostFilter{PostIDs: req.PostIDs})
	if err != nil || len(posts) == 0 {
		c.JSON(http.StatusBadRequest, "post not found")
		return
	}
	c.JSON(http.StatusOK, posts)
}

func (h Handler) GetRankedPosts(c *gin.Context) {
	req := vo.GetRankedPostsFilter{}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	// TODO: pull out set context to middleware
	ctx := po.SetDbInContext(c.Request.Context(), db)
	rdbCtx := po.SetRedisInContext(ctx, redisDb)
	posts, err := h.PostService.GetRankedPosts(rdbCtx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if len(posts) == 0 {
		c.JSON(http.StatusBadRequest, "post not found")
		return
	}
	c.JSON(http.StatusOK, posts)
}

func (h Handler) CreatePost(c *gin.Context) {
	newPost := po.Post{}
	if err := c.BindJSON(&newPost); err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	// TODO: pull out set context to middleware
	ctx := po.SetDbInContext(c.Request.Context(), db)
	res := make(chan vo.PostResp)
	h.PostService.Create(ctx, &newPost, res)
	createResp := <-res
	if createResp.Error != nil {
		c.JSON(http.StatusInternalServerError, createResp.Error)
		return
	}
	c.JSON(http.StatusCreated, createResp.Post)
}

func (h Handler) CreateComment(c *gin.Context) {
	newComment := po.Comment{}
	if err := c.BindJSON(&newComment); err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	// TODO: pull out set context to middleware
	ctx := po.SetDbInContext(c.Request.Context(), db)
	comment, err := h.PostService.CreateComment(ctx, &newComment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}
