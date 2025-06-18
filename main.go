//go:build !embed

package main

import (
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/MhunterDev/img2ascii/source/handlers"
	"github.com/gin-gonic/gin"
)

var (
	outputDir     = getEnv("IMG2ASCII_OUTPUT_DIR", "/tmp/img2ascii")
	outputFile    = getEnv("IMG2ASCII_OUTPUT_FILE", "/tmp/img2ascii/output.txt")
	wwwDir        = getEnv("IMG2ASCII_WWW_DIR", "/tmp/img2ascii/www")
	maxUploadSize = int64(2 << 20)
	maxBannerLen  = 64
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func checkAndPopulate() error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0700); err != nil {
			log.Printf("Failed to create outputDir: %v", err)
			return err
		}
	}
	if _, err := os.Stat(wwwDir); os.IsNotExist(err) {
		if err := os.Mkdir(wwwDir, 0700); err != nil {
			log.Printf("Failed to create wwwDir: %v", err)
			return err
		}
	}
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		f, err := os.Create(outputFile)
		if err != nil {
			log.Printf("Failed to create outputFile: %v", err)
			return err
		}
		f.Close()
	}
	return nil
}

func allowedFileType(header *multipart.FileHeader) bool {
	switch header.Header.Get("Content-Type") {
	case "image/png", "image/jpeg", "image/gif":
		return true
	default:
		return false
	}
}

var globalTmpl *template.Template

func getStaticFS() (tmpl *template.Template, staticFS http.FileSystem, err error) {
	tmpl, err = template.ParseFiles("source/www/index.html")
	if err != nil {
		return nil, nil, err
	}
	return tmpl, http.Dir(wwwDir), nil
}

func main() {
	if err := checkAndPopulate(); err != nil {
		log.Fatalf("Startup error: %v", err)
	}

	tmpl, staticFS, err := getStaticFS()
	if err != nil {
		log.Fatalf("Static/template error: %v", err)
	}
	globalTmpl = tmpl

	cfg := &handlers.Config{
		OutputDir:     outputDir,
		MaxUploadSize: maxUploadSize,
		MaxBannerLen:  maxBannerLen,
		GlobalTmpl:    globalTmpl,
	}

	r := gin.Default()
	r.StaticFS("/static", staticFS)
	r.GET("/", handlers.HandleHome(cfg))
	r.POST("/upload", handlers.HandleUpload(cfg))
	r.POST("/banner", handlers.HandleBanner(cfg))

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
