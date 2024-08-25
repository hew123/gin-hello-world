package main

import (
	"gin-hello-world/po"
	"gin-hello-world/vo"
	"net/http"
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
	postService := vo.NewPostService(DbName, PostCreateInterval)
	handler := Handler{postService}

	// TODO: add auth middleware
	router.GET("/post/get", handler.GetPosts)
	router.POST("/post/create", handler.CreatePost)
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
	posts, err := h.PostService.Find(po.FindPostFilter{PostIDs: req.PostIDs})
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
	res := make(chan vo.PostResp)
	h.PostService.Create(&newPost, res)
	createResp := <-res
	if createResp.Error != nil {
		c.JSON(http.StatusInternalServerError, createResp.Error)
		return
	}
	c.JSON(http.StatusCreated, createResp.Post)
}
