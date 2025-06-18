package banners

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/MhunterDev/img2ascii/source/img2ascii"
	"github.com/fogleman/gg"
	xdraw "golang.org/x/image/draw"
)

const fontRoot = "source/banners/fonts"

type Font string

func (f Font) Path() string {
	return filepath.Join(fontRoot, string(f)+".ttf")
}

type BannerOptions struct {
	Font       Font
	Reverse    bool
	Characters string
	Style      string // e.g. "bold", "italic", etc.
}

type Banner struct {
	Message string
	Path    string // Path to the image file
	Width   int    // in characters
	Height  int    // in lines
	Options BannerOptions
}

// RenderToImage renders the banner message to a PNG image and saves it to outPath.
func (b *Banner) renderToImage() (string, error) {
	outPath := b.Path
	if b.Width <= 0 {
		b.Width = 80 // More readable default width
	}
	if b.Height <= 0 {
		b.Height = 10 // More readable default height
	}
	imgWidth := 5 + (b.Width * 16)   // 16px per character (approx)
	imgHeight := 5 + (b.Height * 16) // 16px per line (approx)
	dc := gg.NewContext(imgWidth, imgHeight)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	fontSize := float64(imgHeight) * 0.8 // Larger font for clarity
	fontPath := b.Options.Font.Path()
	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		return "", fmt.Errorf("failed to load font: %w", err)
	}
	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored(b.Message, float64(imgWidth)/2, float64(imgHeight)/2, 0.5, 0.5)
	pngPath := outPath + ".png"
	if err := dc.SavePNG(pngPath); err != nil {
		return "", err
	}
	return pngPath, nil
}

// resizeRGBA resizes an RGBA image to the target width and height using bilinear scaling.
func (b *Banner) resizeRGBA(pngPath string) (*image.RGBA, error) {
	src, err := gg.LoadPNG(pngPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %v", err)
	}
	dst := image.NewRGBA(image.Rect(0, 0, b.Width, b.Height))
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst, nil
}

func RenderBanner(b Banner) error {
	pngPath, err := b.renderToImage()
	if err != nil {
		return fmt.Errorf("failed to render banner: %w", err)
	}
	resizedImg, err := b.resizeRGBA(pngPath)
	if err != nil {
		return fmt.Errorf("failed to resize banner image: %w", err)
	}
	resizedPngPath := pngPath + ".resized.png"
	if err := gg.SavePNG(resizedPngPath, resizedImg); err != nil {
		return fmt.Errorf("failed to save resized image: %w", err)
	}
	asciiPath := b.Path + ".txt"
	err = img2ascii.RunBanner(resizedPngPath, asciiPath, b.Width, b.Height) // Convert the image to ASCII
	if err != nil {
		return fmt.Errorf("failed to convert image to ASCII: %w", err)
	}
	return nil
}

// In img2ascii.go, update bannerASCII to use a high-contrast set
// func bannerASCII(mode bool, i int) string {
// 	ascii := "@#*+=-:. "
// 	if !mode {
// 		a := reverseString(ascii)
// 		idx := i * (len(ascii) - 1) / 255
// 		return string(a[idx])
// 	}
// 	idx := i * (len(ascii) - 1) / 255
// 	return string(ascii[idx])
// }
