package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"image/jpeg"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/icemint0828/imgedit"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func main() {
	dir, trim, rec, cmp := getArgs()
	fmt.Printf("getArgs dir: %v, trim: %v, rec: %v, cmp: %v\n", dir, trim, rec, cmp)
	if rec {
		files, err := os.ReadDir(dir)
		if err != nil {
			panic(err)
		}
		fmt.Printf("read dir. count: %v\n", len(files))
		for _, file := range files {
			if file.IsDir() {
				createPdf(filepath.Join(dir, file.Name()), trim, cmp)
			}
		}
	}	else {
		createPdf(dir, trim, cmp)
	}
	fmt.Printf("finished\n")
}

func getArgs() (dir, trim string, rec bool, cmp int) {
	flag.Parse()
	args := flag.Args()
	rec, err := strconv.ParseBool(args[2])
	if err != nil {
		panic(err)
	}
	cmp, err = strconv.Atoi(args[3])
	if err != nil {
		panic(err)
	}
	return args[0], args[1], rec, cmp
}

func createPdf(dir, trim string, cmp int) {
	fragments := strings.Split(dir, string(filepath.Separator))
	filename := fragments[len(fragments)-1]
	fmt.Printf("starting... filename: %v\n", filename)
	fmt.Printf("getting images... dir: %v\n", dir)
	originalPaths := getImages(dir)
	fmt.Printf("got images. count: %v\n", len(originalPaths))
	fmt.Printf("copying images to temp dir... trim: %v\n", trim)
	tmpDir, copiedPaths := copyToTemp(&originalPaths, trim)
	fmt.Printf("copied images to temp dir. count: %v, tmpDir: %v\n", len(copiedPaths), tmpDir)
	fmt.Printf("compressing images... cmp: %v percent\n", cmp)
	compressedPaths := compImages(tmpDir, &copiedPaths, cmp)
	fmt.Printf("compressed images... count: %v\n", len(compressedPaths))
	fmt.Printf("appending images to pdf... filename: %v\n", filename)
	appendImagesToPdf(filename, &compressedPaths)
	fmt.Printf("imported images file: %v\n", filename)
	deleteTemp(tmpDir)
}

func getImages(dir string) []string {
	var paths []string
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".png" || filepath.Ext(file.Name()) == ".PNG" ||
			filepath.Ext(file.Name()) == ".jpg" || filepath.Ext(file.Name()) == ".jpeg" {
			paths = append(paths, filepath.Join(dir, file.Name()))
		}
	}
	return paths
}

func copyToTemp(paths *[]string, trim string) (string, []string) {
	var newPaths []string
	tmpDir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(*paths))
	// TODO: create batches
	for _, path := range *paths {
		go func(path string) {
			defer wg.Done()

			copyOrTrimImg(tmpDir, &newPaths, path, trim)
		}(path)
	}
	wg.Wait()
	return tmpDir, newPaths
}

func copyOrTrimImg(tmpDir string, newPaths *[]string, oldPath, trim string) {
	fc, _, err := imgedit.NewFileConverter(oldPath)
	if err != nil {
		panic(err)
	}

	size := fc.Convert().Bounds().Size()
	if size.X > size.Y {
		if trim == "r2l" {
			fc.Trim(size.X/2, 0, size.X/2, size.Y)
		} else if trim == "l2r" {
			fc.Trim(0, 0, size.X/2, size.Y)
		}
		bases1 := strings.Split(filepath.Base(oldPath), ".")
		filename1 := bases1[0] + "_1.png"
		newPath1 := filepath.Join(tmpDir, filename1)
		err = fc.SaveAs(newPath1, imgedit.Png)
		if err != nil {
			panic(err)
		}

		afc, _, err := imgedit.NewFileConverter(oldPath)
		if err != nil {
			panic(err)
		}
		if trim == "r2l" {
			afc.Trim(0, 0, size.X/2, size.Y)
		} else if trim == "l2r" {
			afc.Trim(size.X/2, 0, size.X/2, size.Y)
		}
		bases2 := strings.Split(filepath.Base(oldPath), ".")
		filename2 := bases2[0] + "_2.png"
		newPath2 := filepath.Join(tmpDir, filename2)
		err = afc.SaveAs(newPath2, imgedit.Png)
		if err != nil {
			panic(err)
		}
		*newPaths = append(*newPaths, newPath1, newPath2)
	} else {
		newPath := filepath.Join(tmpDir, filepath.Base(oldPath))
		err = fc.SaveAs(newPath, imgedit.Png)
		if err != nil {
			panic(err)
		}
		*newPaths = append(*newPaths, newPath)
	}
}

func compImages(tmpDir string, paths *[]string, cmp int) []string {
	var newPaths []string

	var wg sync.WaitGroup
	wg.Add(len(*paths))
	// TODO: create batches
	for _, path := range *paths {
		go func(path string) {
			defer wg.Done()

			newPath := compImage(tmpDir, path, cmp)
			newPaths = append(newPaths, newPath)
		}(path)
	}
	wg.Wait()
	return newPaths
}

func compImage(tmpDir, path string, cmp int) string {
	bases := strings.Split(filepath.Base(path), ".")
	filename := bases[0] + "_compressed.jpg"
	newPath := filepath.Join(tmpDir, filename)

	var inFile *os.File
	var outFile *os.File
	var img image.Image
	var err error
	if inFile, err = os.Open(path); err != nil {
		panic(err)
	}
	defer inFile.Close()
	if img, err = png.Decode(inFile); err != nil {
		panic(err)
	}
	if outFile, err = os.Create(newPath); err != nil {
		panic(err)
	}
	defer outFile.Close()
	var option *jpeg.Options
	if cmp < 100 {
		option = &jpeg.Options{Quality: cmp}
	}
	if err = jpeg.Encode(outFile, img, option); err != nil {
		panic(err)
	}
	return newPath
}

func appendImagesToPdf(filename string, paths *[]string) {
	sort.Strings(*paths)

	// https://pkg.go.dev/github.com/pdfcpu/pdfcpu/pkg/api
	// Import images by creating an A4 page for each image.
	// Images are page centered with 1.0 relative scaling.
	// Import an image as a new page of the existing out.pdf.
	imp, _ := api.Import("form:A4, pos:c, s:1.0", types.POINTS)
	api.ImportImagesFile(*paths, filename + ".pdf", imp, nil)
}

func deleteTemp(tmpDir string) {
	if err := os.RemoveAll(tmpDir); err != nil {
		panic(err)
	}
}
