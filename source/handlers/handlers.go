package handlers

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	// First check MIME type
	contentType := header.Header.Get("Content-Type")
	switch contentType {
	case "image/png", "image/jpeg", "image/gif":
		// MIME type is allowed, continue with file signature validation
	default:
		return false
	}

	// Open and read file signature to validate actual content
	file, err := header.Open()
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 512 bytes for signature detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}
	buffer = buffer[:n]

	// Check file signatures (magic numbers)
	if len(buffer) < 4 {
		return false
	}

	// PNG signature: 89 50 4E 47
	if bytes.HasPrefix(buffer, []byte{0x89, 0x50, 0x4E, 0x47}) {
		return true
	}

	// JPEG signature: FF D8 FF
	if bytes.HasPrefix(buffer, []byte{0xFF, 0xD8, 0xFF}) {
		return true
	}

	// GIF signature: GIF87a or GIF89a
	if bytes.HasPrefix(buffer, []byte("GIF87a")) || bytes.HasPrefix(buffer, []byte("GIF89a")) {
		return true
	}

	return false
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
			log.Printf("File upload error: %v", err)
			c.String(400, "File upload failed")
			return
		}

		if upFile.Size > cfg.MaxUploadSize {
			log.Printf("File too large: %d bytes (max: %d)", upFile.Size, cfg.MaxUploadSize)
			c.String(400, "File too large")
			return
		}

		if !AllowedFileType(upFile) {
			log.Printf("Unsupported file type: %s", upFile.Header.Get("Content-Type"))
			c.String(400, "Unsupported file type")
			return
		}

		// Create safe temporary file with sanitized name
		safeFilename := sanitizeFilename(upFile.Filename)
		tmpFile, err := os.CreateTemp("/tmp", "img2ascii_"+safeFilename+"_*.tmp")
		if err != nil {
			log.Printf("Failed to create temp file: %v", err)
			c.String(500, "Internal server error")
			return
		}

		defer func() {
			tmpFile.Close()
			if err := os.Remove(tmpFile.Name()); err != nil {
				log.Printf("Failed to remove temp file: %v", err)
			}
		}()

		file, err := upFile.Open()
		if err != nil {
			log.Printf("Failed to open uploaded file: %v", err)
			c.String(500, "Internal server error")
			return
		}
		defer file.Close()

		// Limit the amount of data we'll copy to prevent DoS
		limitedReader := io.LimitReader(file, cfg.MaxUploadSize)
		if _, err := io.Copy(tmpFile, limitedReader); err != nil {
			log.Printf("Failed to save uploaded file: %v", err)
			c.String(500, "Internal server error")
			return
		}

		// Ensure data is written to disk
		if err := tmpFile.Sync(); err != nil {
			log.Printf("Failed to sync temp file: %v", err)
			c.String(500, "Internal server error")
			return
		}

		outputID := uuid.New().String()
		outputPath := filepath.Join(cfg.OutputDir, fmt.Sprintf("output-%s.txt", outputID))

		runErr := img2ascii.Run(true, tmpFile.Name(), outputPath)
		if runErr != nil {
			log.Printf("ASCII conversion error: %v", runErr)
			c.String(500, "Conversion failed")
			return
		}

		defer func() {
			if err := os.Remove(outputPath); err != nil {
				log.Printf("Failed to remove output file: %v", err)
			}
		}()

		if _, err := os.Stat(outputPath); err != nil {
			log.Printf("Output file not found: %v", err)
			c.String(500, "Conversion failed")
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

		// Sanitize and validate input
		cleanText, err := sanitizeBannerText(bannerText)
		if err != nil {
			log.Printf("Invalid banner text: %v", err)
			c.String(400, "Invalid banner text")
			return
		}

		if len(cleanText) > cfg.MaxBannerLen {
			c.String(400, "Banner text too long")
			return
		}

		outputID := uuid.New().String()
		outputPath := filepath.Join(cfg.OutputDir, fmt.Sprintf("banner-%s", outputID))

		banner := banners.Banner{
			Message: cleanText,
			Path:    outputPath,
			Width:   50,
			Height:  15,
			Options: banners.BannerOptions{
				Font:    "Notable-Regular",
				Reverse: true,
			},
		}

		if err := banners.RenderBanner(banner); err != nil {
			log.Printf("Banner generation error: %v", err)
			c.String(500, "Banner generation failed")
			return
		}

		asciiPath := outputPath + ".txt"
		data, err := os.ReadFile(asciiPath)
		if err != nil {
			log.Printf("Failed to read ASCII output: %v", err)
			c.String(500, "Failed to read banner output")
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

// Input validation and sanitization helpers
var (
	// Allow alphanumeric, spaces, basic punctuation for banner text
	allowedBannerChars = regexp.MustCompile(`^[a-zA-Z0-9\s\.,!?\-_]+$`)
)

// sanitizeBannerText cleans and validates banner text input
func sanitizeBannerText(input string) (string, error) {
	// Trim whitespace
	cleaned := strings.TrimSpace(input)

	// Check length
	if len(cleaned) == 0 {
		return "", fmt.Errorf("banner text cannot be empty")
	}

	// Check for allowed characters
	if !allowedBannerChars.MatchString(cleaned) {
		return "", fmt.Errorf("banner text contains invalid characters")
	}

	// HTML escape for additional safety
	cleaned = html.EscapeString(cleaned)

	return cleaned, nil
}

// sanitizeFilename ensures uploaded filename is safe
func sanitizeFilename(filename string) string {
	// Handle empty filename first
	if strings.TrimSpace(filename) == "" {
		return "upload_" + uuid.New().String()[:8]
	}

	// Get base filename without path
	base := filepath.Base(filename)

	// Handle case where filepath.Base returns "."
	if base == "." || base == ".." {
		return "upload_" + uuid.New().String()[:8]
	}

	// Remove any potentially dangerous characters
	safe := regexp.MustCompile(`[^a-zA-Z0-9\.\-_]`).ReplaceAllString(base, "_")

	// Ensure it's not empty and has reasonable length
	if len(safe) == 0 || len(safe) > 255 {
		safe = "upload_" + uuid.New().String()[:8]
	}

	return safe
}
