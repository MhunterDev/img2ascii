package main

import (
	"html/template"
	"os"
	"os/exec"
	"sync"

	"github.com/MhunterDev/img2ascii/source/img2ascii"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var tmpl *template.Template

const (
	outputDir  = "/tmp/img2ascii"
	outputFile = "/tmp/img2ascii/output.txt"
	wwwDir     = "/tmp/img2ascii/www"
)

func checkAndPopulate() error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0700); err != nil {
			panic(err)
		}
	}
	if _, err := os.Stat(wwwDir); os.IsNotExist(err) {
		if err := os.Mkdir(wwwDir, 0700); err != nil {
			panic(err)
		}

	}
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		f, err := os.Create(outputFile)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
	return nil
}

func handleUpload(c *gin.Context) {
	upFile, err := c.FormFile("file")
	if err != nil {
		c.String(400, "File upload error: %v", err)
		return
	}
	tmpFile, err := os.CreateTemp("/tmp", "upload-*.img")
	if err != nil {
		c.String(500, "Failed to create temp file: %v", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	file, err := upFile.Open()
	if err != nil {
		c.String(500, "Failed to open uploaded file: %v", err)
		return
	}
	defer file.Close()
	if _, err := tmpFile.ReadFrom(file); err != nil {
		c.String(500, "Failed to save uploaded file: %v", err)
		return
	}

	outputID := uuid.New().String()
	outputPath := "/tmp/img2ascii/output-" + outputID + ".txt"
	var wg sync.WaitGroup
	wg.Add(1)
	var runErr error
	go func() {
		defer wg.Done()
		runErr = img2ascii.Run(true, tmpFile.Name(), outputPath)
	}()
	wg.Wait()
	if runErr != nil {
		c.String(500, "ASCII conversion error: %v", runErr)
		return
	}
	defer os.Remove(outputPath)
	if _, err := os.Stat(outputPath); err != nil {
		c.String(500, "Output file not found: %v", err)
		return
	}
	c.File(outputPath)
}

func handleHome(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	err := tmpl.Execute(c.Writer, nil)
	if err != nil {
		c.String(500, "Template execution error: %v", err)
	}
}

func main() {

	if err := checkAndPopulate(); err != nil {
		panic(err)
	}

	script := "cp source/www/index.html /tmp/img2ascii/www/index.html && cp source/www/styles.css /tmp/img2ascii/www/styles.css && cp source/www/main.js /tmp/img2ascii/www/main.js"
	if err := exec.Command("sh", "-c", script).Run(); err != nil {
		panic(err)
	}

	var err error
	tmpl, err = template.ParseFiles("source/www/index.html")
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Static("/static", wwwDir)
	r.GET("/", handleHome)
	r.POST("/upload", handleUpload)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
