package main

import (
	"os"
	"sync"

	"github.com/MhunterDev/img2ascii/source/img2ascii"
	"github.com/gin-gonic/gin"
)

func handleUpload(c *gin.Context) {
	// Parse the multipart form
	upFile, err := c.FormFile("file")
	if err != nil {
		c.String(400, "File upload error: %v", err)
		return
	}
	// Save the uploaded file to a temporary location
	if err := c.SaveUploadedFile(upFile, upFile.Filename); err != nil {
		c.String(500, "Failed to save uploaded file: %v", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		img2ascii.Run(true, upFile.Filename)
	}()
	wg.Wait()

	c.File("/tmp/img2ascii.txt")
}

func main() {
	file, err := os.MkdirTemp("/tmp", "img2ascii")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(file)
	r := gin.Default()

	r.POST("/upload", handleUpload)

	// Serve static files from the "static" directory
	r.Static("/static", "./static")

	// Serve the HTML file
	r.LoadHTMLFiles("./static/index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// Start the server
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
