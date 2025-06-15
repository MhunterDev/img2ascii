package main

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

func NewImage(imgPath string, targetWidth, targetHeight int) (*Image, error) {
	// Open the image file
	file, err := os.Open(imgPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	d, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(bytes.NewReader(d))
	if err != nil {
		// handle error
		panic(err)
	}

	bounds := img.Bounds()
	rgbaImg := image.NewRGBA(bounds)
	imagedraw.Draw(rgbaImg, bounds, img, bounds.Min, imagedraw.Src)

	width := bounds.Dx()
	height := bounds.Dy()

	// Resize if needed
	if targetWidth > 0 && targetHeight > 0 {
		rgbaImg = resizeRGBA(rgbaImg, targetWidth, targetHeight)
		width = targetWidth
		height = targetHeight
	}

	return &Image{
		Name: imgPath,
		Res:  Resolution{Width: width, Height: height},
		Data: rgbaImg.Pix, // Use RGBA pixel data
	}, nil
}

type Pixel struct {
	R int
	G int
	B int
	A int
}

func (r Resolution) PixelCount() int {
	// Calculate the total number of pixels in the image
	return r.Width * r.Height
}

func (p Pixel) CalculateLuminance() int {
	// Using the formula for perceived luminance
	l := 0.2126*float64(p.R) + 0.7152*float64(p.G) + 0.0722*float64(p.B)
	return int(math.Round(l))

}

func (i Image) ToLumScores() []int {
	// Convert the image data to a slice of Pixel structs
	pixels := make([]Pixel, i.Res.PixelCount())
	for j := 0; j < i.Res.Height; j++ {
		for k := 0; k < i.Res.Width; k++ {
			idx := (j*i.Res.Width + k) * 4 // Assuming RGBA format
			if idx+3 < len(i.Data) {
				pixels[j*i.Res.Width+k] = Pixel{
					R: int(i.Data[idx]),
					G: int(i.Data[idx+1]),
					B: int(i.Data[idx+2]),
					A: int(i.Data[idx+3]),
				}
			}
		}
	}
	lScores := make([]int, len(pixels))
	for idx, pixel := range pixels {
		lScores[idx] = pixel.CalculateLuminance()
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
	ascii := "@#X%&$*+=o-:,. "
	if !mode {
		a := reverseString(ascii)
		idx := i * (len(ascii) - 1) / 255
		return string(a[idx])
	}
	idx := i * (len(ascii) - 1) / 255

	return string(ascii[idx])
}

func (i Image) ToASCII(mode bool) string {
	// Convert the image to an ASCII representation
	lScores := i.ToLumScores()
	asciiArt := ""
	for j := 0; j < i.Res.Height; j++ {
		for k := 0; k < i.Res.Width; k++ {
			idx := j*i.Res.Width + k
			if idx < len(lScores) {
				asciiArt += makeASCII(mode, lScores[idx])
			}
		}
		asciiArt += "\n"
	}
	return asciiArt
}

// Resize RGBA image to fit target width and height
func resizeRGBA(src *image.RGBA, targetWidth, targetHeight int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

func main() {
	imgPath := "/home/matth/Downloads/dryRun.png"

	// Open image to get original dimensions
	file, err := os.Open(imgPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Desired output height (rows)
	targetHeight := 75

	// Correction factor for ASCII aspect ratio
	aspectCorrection := 0.5 // Try 0.5 or 0.55 for best results

	// Calculate target width based on aspect ratio and correction
	targetWidth := int(float64(origWidth) * (float64(targetHeight) / float64(origHeight)) / aspectCorrection)

	imgObj, err := NewImage(imgPath, targetWidth, targetHeight)
	if err != nil {
		panic(err)
	}

	asciiArt := imgObj.ToASCII(true)
	fileOut, err := os.OpenFile("/tmp/ascii_art.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer fileOut.Close()
	_, err = fileOut.WriteString(asciiArt)
	if err != nil {
		panic(err)
	}
}
