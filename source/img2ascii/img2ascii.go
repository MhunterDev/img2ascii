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

type Resolution struct {
	Width  int
	Height int
}

type Image struct {
	Name string
	Res  Resolution
	Data []byte
}

func Run(mode bool, imgPath string, outputPath string) error {
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
	asciiArt := imgObj.toASCII(mode)
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

func makeASCII(mode bool, i int) string {
	ascii := "@#%*o()1l=:-."
	if !mode {
		a := reverseString(ascii)
		idx := i * (len(ascii) - 1) / 255
		return string(a[idx])
	}
	idx := i * (len(ascii) - 1) / 255
	return string(ascii[idx])
}

func (i Image) toASCII(mode bool) string {
	lScores := i.toLumScores()
	var asciiArt strings.Builder
	asciiArt.Grow(i.Res.Width * (i.Res.Height + 1))
	for j := 0; j < i.Res.Height; j++ {
		for k := 0; k < i.Res.Width; k++ {
			idx := j*i.Res.Width + k
			if idx < len(lScores) {
				asciiArt.WriteString(makeASCII(mode, lScores[idx]))
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

	asciiArt := imgObj.toASCIIBanner(true)
	fileOut, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fileOut.Close()
	_, err = fileOut.WriteString(asciiArt)
	_ = os.WriteFile("img2ascii.log", []byte(asciiArt), 0644)
	return err
}

func (i Image) toASCIIBanner(mode bool) string {
	lScores := i.toLumScores()
	var asciiArt strings.Builder
	asciiArt.Grow(i.Res.Width * (i.Res.Height + 1))
	for j := 0; j < i.Res.Height; j++ {
		for k := 0; k < i.Res.Width; k++ {
			idx := j*i.Res.Width + k
			if idx < len(lScores) {
				asciiArt.WriteString(bannerASCII(mode, lScores[idx]))
			}
		}
		asciiArt.WriteByte('\n')
	}
	return asciiArt.String()
}

func bannerASCII(mode bool, i int) string {
	ascii := "@#*+=-:. "
	if !mode {
		a := reverseString(ascii)
		idx := i * (len(ascii) - 1) / 255
		return string(a[idx])
	}
	idx := i * (len(ascii) - 1) / 255
	return string(ascii[idx])
}
