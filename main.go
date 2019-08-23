package main

import (
	"fmt"
	"github.com/otiai10/marmoset"
	"github.com/otiai10/ocrserver/filters"
	"log"
	"net/http"
	"os"

	"github.com/otiai10/ocrserver/controllers"
)

var logger *log.Logger

func main() {

	marmoset.LoadViews("./app/views")

	r := marmoset.NewRouter()

	//设置访问的路由
	// API
	r.GET("/status", controllers.Status)
	r.Handle("/base64", &controllers.Base64Controller{})
	r.Handle("/file", &controllers.FileController{})

	//r.POST("/file", controllers.FileUpload)
	// Sample Page
	r.GET("/", controllers.Index)
	r.Static("/assets", "./app/assets")

	logger = log.New(os.Stdout, fmt.Sprintf("[%s] ", "ocrserver"), 0)
	r.Apply(&filters.LogFilter{Logger: logger})
	r.Apply(&filters.SignFilter{})

	port := "8080"
	if port == "" {
		logger.Fatalln("Required env `PORT` is not specified.")
	}
	logger.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Println(err)
	}
}
