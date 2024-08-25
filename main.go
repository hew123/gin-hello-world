package main

import (
	"context"
	"gin-hello-world/po"
	"gin-hello-world/vo"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

func main() {
	router := gin.Default()

	dbConn, err := gorm.Open(sqlite.Open(DbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db = dbConn
	ctx := po.SetDbInContext(context.Background(), dbConn)
	postService := vo.NewPostService(ctx, PostCreateInterval)
	handler := Handler{postService}

	// TODO: add auth middleware
	router.GET("/post/get", handler.GetPosts)
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
