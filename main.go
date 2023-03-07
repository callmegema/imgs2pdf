package main

import (
	"flag"
	"fmt"
	"os"
	// "io"
	"io/ioutil"
  "path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/icemint0828/imgedit"
)

const TempDir = "tmp"

func main() {
	dir, filename := getArgs()
	fmt.Printf("get dir: %v\n", dir)
	paths := getImages(dir)
	fmt.Printf("get images. count: %v\n", len(paths))
	newPaths := copyToTemp(paths)
	fmt.Printf("copied images to temp dir. count: %v\n", len(newPaths))
	createPdf(filename, newPaths)
	fmt.Printf("created pdf: %v\n", filename)
	deleteTemp()
	fmt.Printf("finished\n")
}

func getArgs() (dir string, filename string) {
	flag.Parse()
	args := flag.Args()
	dir = args[0]
	fragments := strings.Split(dir, string(filepath.Separator))
	filename = fragments[len(fragments)-1]
	return dir, filename
}

func getImages(dir string) []string {
	var paths []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	// TODO: use goroutine
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".png" || filepath.Ext(file.Name()) == ".PNG" ||
			filepath.Ext(file.Name()) == ".jpg" || filepath.Ext(file.Name()) == ".jpeg" {
			paths = append(paths, filepath.Join(dir, file.Name()))
		}
	}
	return paths
}

func copyToTemp(paths []string) []string {
	var newPaths []string
	if err := os.RemoveAll(TempDir); err != nil {
		panic(err)
	}
	if err := os.Mkdir(TempDir, 0777); err != nil {
		panic(err)
	}
	for _, path := range paths {
		newPaths = copyOrTrimImg(newPaths, path)
	}
	return newPaths
}

func copyOrTrimImg(newPaths []string, oldPath string) []string {
	fc, _, err := imgedit.NewFileConverter(oldPath)
	if err != nil {
		panic(err)
	}

	size := fc.Convert().Bounds().Size()
	if size.X > size.Y {
		fc.Trim(size.X/2, 0, size.X/2, size.Y)
		bases1 := strings.Split(filepath.Base(oldPath), ".")
		filename1 := bases1[0] + "_1.png"
		newPath1 := filepath.Join(TempDir, filename1)
		err = fc.SaveAs(newPath1, imgedit.Png)
		if err != nil {
			panic(err)
		}
		newPaths = append(newPaths, newPath1)

		afc, _, err := imgedit.NewFileConverter(oldPath)
		if err != nil {
			panic(err)
		}
		afc.Trim(0, 0, size.X/2, size.Y)
		bases2 := strings.Split(filepath.Base(oldPath), ".")
		filename2 := bases2[0] + "_2.png"
		newPath2 := filepath.Join(TempDir, filename2)
		err = afc.SaveAs(newPath2, imgedit.Png)
		if err != nil {
			panic(err)
		}
		newPaths = append(newPaths, newPath2)
	} else {
		newPath := filepath.Join(TempDir, filepath.Base(oldPath))
		err = fc.SaveAs(newPath, imgedit.Png)
		if err != nil {
			panic(err)
		}
		newPaths = append(newPaths, newPath)
	}
	return newPaths
}

func createPdf(filename string, paths []string) {
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
