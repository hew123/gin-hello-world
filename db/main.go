package main

import (
	"fmt"
	"gin-hello-world/po"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&po.Post{})
	db.AutoMigrate(&po.Comment{})
	//db.Create(&po.Post{ID: 11, Caption: "hello world"})

	res := po.Post{}
	db.Find(&res, 11)
	fmt.Println("result: ", res)
}
