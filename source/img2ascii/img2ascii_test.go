package img2ascii

import (
	"image"
	"image/color"
	"testing"
)

func TestCalculateLuminance(t *testing.T) {
	tests := []struct {
		name     string
		r, g, b  int
		expected int
	}{
		{"Black", 0, 0, 0, 0},
		{"White", 255, 255, 255, 255},
		{"Red", 255, 0, 0, 54},
		{"Green", 0, 255, 0, 182},
		{"Blue", 0, 0, 255, 18},
		{"Gray", 128, 128, 128, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateLuminance(tt.r, tt.g, tt.b)
			if result != tt.expected {
				t.Errorf("calculateLuminance(%d, %d, %d) = %d, want %d",
					tt.r, tt.g, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMakeASCII(t *testing.T) {
	tests := []struct {
		name      string
		mode      ConversionMode
		reverse   bool
		luminance int
		expected  string
	}{
		{"Default mode, no reverse, dark", ModeDefault, false, 0, "@"},
		{"Default mode, no reverse, light", ModeDefault, false, 255, "."},
		{"Default mode, reverse, dark", ModeDefault, true, 0, "."},
		{"Default mode, reverse, light", ModeDefault, true, 255, "@"},
		{"Banner mode, no reverse, dark", ModeBanner, false, 0, "@"},
		{"Banner mode, no reverse, light", ModeBanner, false, 255, " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeASCII(tt.mode, tt.reverse, tt.luminance)
			if result != tt.expected {
				t.Errorf("makeASCII(%v, %v, %d) = %q, want %q",
					tt.mode, tt.reverse, tt.luminance, result, tt.expected)
			}
		})
	}
}

func TestReverseString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Single character", "a", "a"},
		{"Simple string", "abc", "cba"},
		{"ASCII chars", "@#%*", "*%#@"},
		{"Unicode", "αβγ", "γβα"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reverseString(tt.input)
			if result != tt.expected {
				t.Errorf("reverseString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestResolution_pixelCount(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		expected      int
	}{
		{"1x1", 1, 1, 1},
		{"10x10", 10, 10, 100},
		{"65x54", 65, 54, 3510},
		{"0x0", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Resolution{Width: tt.width, Height: tt.height}
			result := r.pixelCount()
			if result != tt.expected {
				t.Errorf("Resolution{%d, %d}.pixelCount() = %d, want %d",
					tt.width, tt.height, result, tt.expected)
			}
		})
	}
}

func createTestImage(width, height int, fillColor color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fillColor)
		}
	}
	return img
}

func TestImage_toLumScores(t *testing.T) {
	// Create a small test image
	width, height := 2, 2
	testImg := createTestImage(width, height, color.RGBA{R: 128, G: 128, B: 128, A: 255})

	img := &Image{
		Name: "test",
		Res:  Resolution{Width: width, Height: height},
		Data: testImg.Pix,
	}

	scores := img.toLumScores()

	if len(scores) != width*height {
		t.Errorf("Expected %d luminance scores, got %d", width*height, len(scores))
	}

	// All pixels should have the same luminance (gray color)
	expectedLuminance := calculateLuminance(128, 128, 128)
	for i, score := range scores {
		if score != expectedLuminance {
			t.Errorf("Pixel %d: expected luminance %d, got %d", i, expectedLuminance, score)
		}
	}
}

func TestImage_toASCII(t *testing.T) {
	// Create a small test image
	width, height := 2, 2
	testImg := createTestImage(width, height, color.RGBA{R: 0, G: 0, B: 0, A: 255}) // Black

	img := &Image{
		Name: "test",
		Res:  Resolution{Width: width, Height: height},
		Data: testImg.Pix,
	}

	ascii := img.toASCII(ModeDefault, false)

	// Should contain newlines and ASCII characters
	if len(ascii) == 0 {
		t.Error("ASCII output is empty")
	}

	// Count newlines (should be equal to height)
	newlineCount := 0
	for _, char := range ascii {
		if char == '\n' {
			newlineCount++
		}
	}

	if newlineCount != height {
		t.Errorf("Expected %d newlines, got %d", height, newlineCount)
	}
}

func TestRunWithInvalidFile(t *testing.T) {
	err := Run(false, "nonexistent.jpg", "/tmp/output.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestRunBannerWithInvalidFile(t *testing.T) {
	err := RunBanner("nonexistent.jpg", "/tmp/output.txt", 50, 15)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

// Integration test with a real small image
func TestRunIntegration(t *testing.T) {
	// Skip integration test for now as it requires more setup
	t.Skip("Integration test requires PNG encoding setup")
}
