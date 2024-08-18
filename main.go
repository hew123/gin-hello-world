package main

import (
	"fmt"
	"gin-hello-world/po"
	"net/http"

	"github.com/gin-gonic/gin"
)

// config
const (
	DbName = "test.db"
)

func main() {
	router := gin.Default()
	postPersistenceSvc := po.NewPostPersistenceService(DbName)
	handler := Handler{postPersistenceSvc}

	// TODO: add auth middleware
	router.GET("/post/get", handler.GetPosts)
	router.POST("/post/create", handler.CreatePost)
	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type Handler struct {
	postPo po.PostPersistenceService
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
	fmt.Printf("Request: %v", req)
	posts, err := h.postPo.Find(po.FindPostFilter{PostIDs: req.PostIDs})
	if err != nil || len(*posts) == 0 {
		c.JSON(http.StatusBadRequest, "post not found")
		return
	}
	c.JSON(http.StatusOK, *posts)
}

func (h Handler) CreatePost(c *gin.Context) {
	newPost := po.Post{}
	if err := c.BindJSON(&newPost); err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	post, err := h.postPo.Create(&newPost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusCreated, post)
}
