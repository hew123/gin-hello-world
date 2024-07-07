package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var UserPersistence map[uint64]User

func main() {
	router := gin.Default()
	handler := Handler{}
	UserPersistence = map[uint64]User{
		1: {ID: 1, UserName: "hello world"},
	}

	router.GET("/user/get/:id", handler.GetUser)
	router.POST("/user/create", handler.CreateUser)
	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type Handler struct {
}

type User struct {
	ID       uint64 `json:"id"`
	UserName string `json:"username"`
}

type GetUserReq struct {
	ID uint64 `uri:"id" binding:"required"`
}

func (h Handler) GetUser(c *gin.Context) {
	req := GetUserReq{}
	if err := c.BindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	fmt.Printf("Request: %v", req)
	user, ok := UserPersistence[req.ID]
	if !ok {
		c.JSON(http.StatusBadRequest, "user not found")
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h Handler) CreateUser(c *gin.Context) {
	newUser := User{}
	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	UserPersistence[newUser.ID] = newUser
	c.JSON(http.StatusCreated, newUser)
}
