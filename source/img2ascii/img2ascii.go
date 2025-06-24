package img2ascii

import (
	"bytes"
	"image"
	imagedraw "image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"runtime"
	"strings"

	xdraw "golang.org/x/image/draw"
)

// ASCII character sets for different purposes
const (
	defaultASCIIChars = "@#%*o()1l=:-."
	bannerASCIIChars  = "@#*+=-:. "
)

// ConversionMode defines how ASCII conversion should be performed
type ConversionMode int

const (
	ModeDefault ConversionMode = iota
	ModeBanner
)

// AspectRatioMode defines how aspect ratio should be handled
type AspectRatioMode int

const (
	AspectScale AspectRatioMode = iota // Scale maintaining aspect ratio (default)
	AspectPixel                        // 1:1 pixel mapping, no scaling
	AspectFixed                        // Fixed output size, ignore aspect ratio
)

type ConversionOptions struct {
	AspectMode  AspectRatioMode
	FixedWidth  int
	FixedHeight int
	Reverse     bool
	Mode        ConversionMode
}

type Resolution struct {
	Width  int
	Height int
}

type Image struct {
	Name string
	Res  Resolution
	Data []byte
}

func Run(reverse bool, imgPath string, outputPath string) error {
	file, err := os.Open(imgPath)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()
	maxWidth := 65
	maxHeight := 54
	var targetWidth, targetHeight int
	imgAspect := float64(origWidth) / float64(origHeight)
	pageAspect := float64(maxWidth) / float64(maxHeight)
	if imgAspect > pageAspect {
		targetWidth = maxWidth
		targetHeight = int(float64(maxWidth) / imgAspect)
		if targetHeight > maxHeight {
			targetHeight = maxHeight
		}
	} else {
		targetHeight = maxHeight
		targetWidth = int(float64(maxHeight) * imgAspect)
		if targetWidth > maxWidth {
			targetWidth = maxWidth
		}
	}
	imgObj, err := newImage(imgPath, targetWidth, targetHeight)
	if err != nil {
		return err
	}
	asciiArt := imgObj.toASCII(ModeDefault, reverse)
	fileOut, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer fileOut.Close()
	_, err = fileOut.WriteString(asciiArt)
	_ = os.WriteFile("img2ascii.log", []byte(asciiArt), 0644)
	return err
}

func newImage(imgPath string, targetWidth, targetHeight int) (*Image, error) {
	file, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	d, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(d))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	rgbaImg := image.NewRGBA(bounds)
	imagedraw.Draw(rgbaImg, bounds, img, bounds.Min, imagedraw.Src)
	width := bounds.Dx()
	height := bounds.Dy()
	if targetWidth > 0 && targetHeight > 0 {
		rgbaImg = resizeRGBA(rgbaImg, targetWidth, targetHeight)
		width = targetWidth
		height = targetHeight
	}
	return &Image{
		Name: imgPath,
		Res:  Resolution{Width: width, Height: height},
		Data: rgbaImg.Pix,
	}, nil
}

type Pixel struct {
	R int
	G int
	B int
	A int
}

func (r Resolution) pixelCount() int {
	return r.Width * r.Height
}

func calculateLuminance(r, g, b int) int {
	l := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	return int(math.Round(l))
}

func (i Image) toLumScores() []int {
	pixelCount := i.Res.pixelCount()
	lScores := make([]int, pixelCount)
	workers := runtime.NumCPU()
	if workers > pixelCount {
		workers = pixelCount
	}
	tasks := make(chan int, workers)
	done := make(chan struct{}, workers)
	for w := 0; w < workers; w++ {
		go func() {
			for idx := range tasks {
				byteIdx := idx * 4
				if byteIdx+3 < len(i.Data) {
					lScores[idx] = calculateLuminance(int(i.Data[byteIdx]), int(i.Data[byteIdx+1]), int(i.Data[byteIdx+2]))
				}
			}
			done <- struct{}{}
		}()
	}
	for idx := 0; idx < pixelCount; idx++ {
		tasks <- idx
	}
	close(tasks)
	for w := 0; w < workers; w++ {
		<-done
	}
	return lScores
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func makeASCII(mode ConversionMode, reverse bool, luminance int) string {
	var ascii string
	switch mode {
	case ModeBanner:
		ascii = bannerASCIIChars
	default:
		ascii = defaultASCIIChars
	}

	if reverse {
		ascii = reverseString(ascii)
	}

	idx := luminance * (len(ascii) - 1) / 255
	if idx >= len(ascii) {
		idx = len(ascii) - 1
	}
	return string(ascii[idx])
}

func (i Image) toASCII(mode ConversionMode, reverse bool) string {
	lScores := i.toLumScores()
	var asciiArt strings.Builder
	asciiArt.Grow(i.Res.Width * (i.Res.Height + 1))
	for j := 0; j < i.Res.Height; j++ {
		for k := 0; k < i.Res.Width; k++ {
			idx := j*i.Res.Width + k
			if idx < len(lScores) {
				asciiArt.WriteString(makeASCII(mode, reverse, lScores[idx]))
			}
		}
		asciiArt.WriteByte('\n')
	}
	return asciiArt.String()
}

func resizeRGBA(src *image.RGBA, targetWidth, targetHeight int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

func RunBanner(imgPath string, outputPath string, width, height int) error {
	file, err := os.Open(imgPath)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()
	maxWidth := width
	maxHeight := height
	var targetWidth, targetHeight int
	imgAspect := float64(origWidth) / float64(origHeight)
	pageAspect := float64(maxWidth) / float64(maxHeight)
	if imgAspect > pageAspect {
		targetWidth = maxWidth
		targetHeight = int(float64(maxWidth) / imgAspect)
		if targetHeight > maxHeight {
			targetHeight = maxHeight
		}
	} else {
		targetHeight = maxHeight
		targetWidth = int(float64(maxHeight) * imgAspect)
		if targetWidth > maxWidth {
			targetWidth = maxWidth
		}
	}

	imgObj, err := newImage(imgPath, targetWidth, targetHeight)
	if err != nil {
		return err
	}

	asciiArt := imgObj.toASCII(ModeBanner, false)
	fileOut, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fileOut.Close()
	_, err = fileOut.WriteString(asciiArt)
	_ = os.WriteFile("img2ascii.log", []byte(asciiArt), 0644)
	return err
}

func RunWithOptions(imgPath string, outputPath string, options ConversionOptions) error {
	file, err := os.Open(imgPath)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	var targetWidth, targetHeight int

	switch options.AspectMode {
	case AspectPixel:
		// 1:1 pixel mapping - use original dimensions with reasonable limits
		targetWidth = origWidth
		targetHeight = origHeight
		// Limit to prevent browser crashes with very large images
		maxPixelWidth := 300
		maxPixelHeight := 200
		if targetWidth > maxPixelWidth {
			ratio := float64(maxPixelWidth) / float64(targetWidth)
			targetWidth = maxPixelWidth
			targetHeight = int(float64(targetHeight) * ratio)
		}
		if targetHeight > maxPixelHeight {
			ratio := float64(maxPixelHeight) / float64(targetHeight)
			targetHeight = maxPixelHeight
			targetWidth = int(float64(targetWidth) * ratio)
		}
	case AspectFixed:
		// Fixed output size - use specified dimensions
		targetWidth = options.FixedWidth
		targetHeight = options.FixedHeight
	default: // AspectScale
		// Scale maintaining aspect ratio (original behavior)
		maxWidth := 65
		maxHeight := 54
		imgAspect := float64(origWidth) / float64(origHeight)
		pageAspect := float64(maxWidth) / float64(maxHeight)
		if imgAspect > pageAspect {
			targetWidth = maxWidth
			targetHeight = int(float64(maxWidth) / imgAspect)
			if targetHeight > maxHeight {
				targetHeight = maxHeight
			}
		} else {
			targetHeight = maxHeight
			targetWidth = int(float64(maxHeight) * imgAspect)
			if targetWidth > maxWidth {
				targetWidth = maxWidth
			}
		}
	}

	imgObj, err := newImage(imgPath, targetWidth, targetHeight)
	if err != nil {
		return err
	}
	asciiArt := imgObj.toASCII(options.Mode, options.Reverse)
	fileOut, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer fileOut.Close()
	_, err = fileOut.WriteString(asciiArt)
	_ = os.WriteFile("img2ascii.log", []byte(asciiArt), 0644)
	return err
}
