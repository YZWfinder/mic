package main

import (
	"fmt"
	"net/http"

	"github.com/YZWfinder/mic"
)

func main() {
	fmt.Println("server start....")

	server := mic.Server()
	//开启静态文件
	server.Public("/static")
	server.Post("/user", func(w http.ResponseWriter, r *http.Request) {
		mic.ServeJson(w, struct {
			Name string `json:"name"`
			Add  string `json:"add"`
		}{Name: "yangzhiwen", Add: "广东深圳"})
	})
	server.Run(":80")
}
