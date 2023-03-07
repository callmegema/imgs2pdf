package main

import (
	"flag"
	"fmt"
	"os"
	"io"
	"io/ioutil"
  "path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/icemint0828/imgedit"
)

const TempDir = "temp"

func main() {
	dir, filename := getArgs()
	fmt.Printf("get dir: %v\n", dir)
	paths := getImages(dir)
	fmt.Printf("get images. count: %v\n", len(paths))
	newPaths := copyToTemp(paths)
	fmt.Printf("copied images to temp dir. count: %v\n", len(newPaths))
	createPdf(filename, newPaths)
	fmt.Printf("created pdf\n")
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
		// TODO: 場合によってTrimする
		newPath := TempDir + "/" + filepath.Base(path)
		newFile, err := os.Create(newPath)
		if err != nil {
			fmt.Println(err)
		}
		oldFile, err := os.Open(path)
		if err != nil {
			fmt.Println(err)
		}
		_, err = io.Copy(newFile, oldFile)
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

func imgconvert() {
	// 切り取り開始位置の指定（px）
	left, top := 100, 100
	// サイズの指定(px)
	width, height := 400, 400

	// FileConverter
	fc, _, err := imgedit.NewFileConverter("/Users/gema/Desktop/カジュアル和食/IMG_6362.PNG")
	if err != nil {
		panic(err)
	}

	// size := fc.Convert().Bounds().Size()
	// fmt.Println(size)

	// トリム
	fc.Trim(left, top, width, height)
	// 保存(jpeg, gif形式での保存も可能)
	err = fc.SaveAs("/Users/gema/Desktop/カジュアル和食/dstImage.png", imgedit.Png)
	if err != nil {
		panic(err)
	}
}
