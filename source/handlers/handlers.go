package handlers

import (
	"html/template"
	"log"
	"mime/multipart"
	"os"
	"strconv"

	"github.com/MhunterDev/img2ascii/source/banners"
	"github.com/MhunterDev/img2ascii/source/img2ascii"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Config struct {
	OutputDir     string
	MaxUploadSize int64
	MaxBannerLen  int
	GlobalTmpl    *template.Template
}

func AllowedFileType(header *multipart.FileHeader) bool {
	switch header.Header.Get("Content-Type") {
	case "image/png", "image/jpeg", "image/gif":
		return true
	default:
		return false
	}
}

func HandleHome(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		if err := cfg.GlobalTmpl.Execute(c.Writer, nil); err != nil {
			c.String(500, "Template execution error: %v", err)
		}
	}
}

func HandleUpload(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		upFile, err := c.FormFile("file")
		if err != nil {
			c.String(400, "File upload error: %v", err)
			return
		}
		if upFile.Size > cfg.MaxUploadSize {
			c.String(400, "File too large (max %s bytes)", strconv.FormatInt(cfg.MaxUploadSize, 10))
			return
		}
		if !AllowedFileType(upFile) {
			c.String(400, "Unsupported file type")
			return
		}
		tmpFile, err := os.CreateTemp("/tmp", "upload-*.img")
		if err != nil {
			log.Printf("Failed to create temp file: %v", err)
			c.String(500, "Failed to create temp file: %v", err)
			return
		}
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				log.Printf("Failed to remove temp file: %v", err)
			}
		}()
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
		outputPath := cfg.OutputDir + "/output-" + outputID + ".txt"
		runErr := img2ascii.Run(true, tmpFile.Name(), outputPath)
		if runErr != nil {
			c.String(500, "ASCII conversion error: %v", runErr)
			return
		}
		defer func() {
			if err := os.Remove(outputPath); err != nil {
				log.Printf("Failed to remove output file: %v", err)
			}
		}()
		if _, err := os.Stat(outputPath); err != nil {
			c.String(500, "Output file not found: %v", err)
			return
		}
		c.File(outputPath)
	}
}

func HandleBanner(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		bannerText := c.PostForm("bannerText")
		if bannerText == "" {
			c.String(400, "No banner text provided")
			return
		}
		if len(bannerText) > cfg.MaxBannerLen {
			c.String(400, "Banner text too long (max %d chars)", cfg.MaxBannerLen)
			return
		}
		outputID := uuid.New().String()
		outputPath := cfg.OutputDir + "/banner-" + outputID
		banner := banners.Banner{
			Message: bannerText,
			Path:    outputPath,
			Width:   50,
			Height:  15,
			Options: banners.BannerOptions{
				Font:    "Notable-Regular",
				Reverse: true,
			},
		}
		if err := banners.RenderBanner(banner); err != nil {
			c.String(500, "Banner generation error: %v", err)
			return
		}
		asciiPath := outputPath + ".txt"
		data, err := os.ReadFile(asciiPath)
		if err != nil {
			c.String(500, "Failed to read ASCII output: %v", err)
			return
		}
		defer func() {
			if err := os.Remove(asciiPath); err != nil {
				log.Printf("Failed to remove banner ascii file: %v", err)
			}
		}()
		c.Data(200, "text/plain; charset=utf-8", data)
	}
}
