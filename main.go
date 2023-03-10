package main

import (
	"flag"
	"fmt"
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

const TempDir = "tmp"

func main() {
	dir, trim, rec := getArgs()
	fmt.Printf("getArgs dir: %v, trim: %v, rec: %v\n", dir, trim, rec)
	if rec {
		files, err := os.ReadDir(dir)
		if err != nil {
			panic(err)
		}
		fmt.Printf("read dir. count: %v\n", len(files))
		for _, file := range files {
			if file.IsDir() {
				createPdf(filepath.Join(dir, file.Name()), trim)
			}
		}
	}	else {
		createPdf(dir, trim)
	}
	fmt.Printf("finished\n")
}

func getArgs() (dir, trim string, rec bool) {
	flag.Parse()
	args := flag.Args()
	rec, err := strconv.ParseBool(args[2])
	if err != nil {
		panic(err)
	}
	return args[0], args[1], rec
}

func createPdf(dir, trim string) {
	fragments := strings.Split(dir, string(filepath.Separator))
	filename := fragments[len(fragments)-1]
	fmt.Printf("got dir: %v, trim: %v, filename: %v\n", dir, trim, filename)
	fmt.Printf("getting images...\n")
	paths := getImages(dir)
	fmt.Printf("got images. count: %v\n", len(paths))
	fmt.Printf("copying images to temp dir...\n")
	newPaths := copyToTemp(paths, trim)
	fmt.Printf("copied images to temp dir. count: %v\n", len(newPaths))
	fmt.Printf("appending images to pdf... filename: %v\n", filename)
	appendImagesToPdf(filename, newPaths)
	fmt.Printf("imported images file: %v\n", filename)
	deleteTemp()
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

func copyToTemp(paths []string, trim string) []string {
	var newPaths []string
	if err := os.RemoveAll(TempDir); err != nil {
		panic(err)
	}
	if err := os.Mkdir(TempDir, 0777); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(paths))
	for _, path := range paths {
		go func(path string) {
			defer wg.Done()

			copyOrTrimImg(&newPaths, path, trim)
		}(path)
	}
	wg.Wait()
	return newPaths
}

func copyOrTrimImg(newPaths *[]string, oldPath, trim string) {
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
		newPath1 := filepath.Join(TempDir, filename1)
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
		newPath2 := filepath.Join(TempDir, filename2)
		err = afc.SaveAs(newPath2, imgedit.Png)
		if err != nil {
			panic(err)
		}
		*newPaths = append(*newPaths, newPath1, newPath2)
	} else {
		newPath := filepath.Join(TempDir, filepath.Base(oldPath))
		err = fc.SaveAs(newPath, imgedit.Png)
		if err != nil {
			panic(err)
		}
		*newPaths = append(*newPaths, newPath)
	}
}

func appendImagesToPdf(filename string, paths []string) {
	sort.Strings(paths)

	// https://pkg.go.dev/github.com/pdfcpu/pdfcpu/pkg/api
	// Import images by creating an A4 page for each image.
	// Images are page centered with 1.0 relative scaling.
	// Import an image as a new page of the existing out.pdf.
	imp, _ := api.Import("form:A4, pos:c, s:1.0", types.POINTS)
	api.ImportImagesFile(paths, filename + ".pdf", imp, nil)
}

func deleteTemp() {
	if err := os.RemoveAll(TempDir); err != nil {
		panic(err)
	}
}
