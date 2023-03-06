package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "example")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test image file in the temporary directory
	err = createTestImage(t, tmpDir)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	// Set up command line arguments
	os.Args = []string{"./main.go", tmpDir, "l2r", "false", "50"}
	main()

	fragments := strings.Split(tmpDir, string(filepath.Separator))
	filename := fragments[len(fragments)-1] + ".pdf"
	if !exists(filename) {
		t.Errorf("Error: expect to have a pdf file: %v\n", filename)
	}
	os.Remove(filename)
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func createTestImage(t *testing.T, dir string) error {
	width := 200
	height := 100
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Set color for each pixel.
	cyan := color.RGBA{100, 200, 200, 0xff}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img.Set(x, y, cyan)
			case x >= width/2 && y >= height/2: // lower right quadrant
				img.Set(x, y, color.White)
			default:
				// Use zero value.
			}
		}
	}

	path := filepath.Join(dir, "test.png")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	png.Encode(f, img)
	return nil
}
