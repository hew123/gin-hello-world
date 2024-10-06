package handler

import (
	"gin-hello-world/vo"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	PostService vo.PostService
}

func NewHandler(vo vo.PostService) Handler {
	return Handler{vo}
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
	posts, err := h.PostService.Find(c, vo.FindPostFilter{PostIDs: req.PostIDs})
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
	posts, err := h.PostService.GetRankedPosts(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, posts)
}

func (h Handler) CreatePost(c *gin.Context) {
	newPost := vo.Post{}
	if err := c.BindJSON(&newPost); err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	post, err := h.PostService.CreatePost(c, newPost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, post)
}

func (h Handler) CreateComment(c *gin.Context) {
	newComment := vo.Comment{}
	if err := c.BindJSON(&newComment); err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	comment, err := h.PostService.CreateComment(c, newComment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}
