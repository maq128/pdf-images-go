package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

type MyImage struct {
	io.Reader
	Name     string
	FileType string
	PageNr   int
	ObjNr    int
}

func main() {
	// 从命令行参数获取 PDF 文件名
	if len(os.Args) < 2 {
		fmt.Println("Usage: extractimages [pdffile]")
		return
	}
	pdfFilename, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	// 打开 PDF 文件
	pdfFile, err := os.Open(pdfFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pdfFile.Close()

	// 创建 ZIP 文件
	zipFilename := strings.TrimSuffix(pdfFilename, ".pdf") + ".images.zip"
	fmt.Println(zipFilename)

	zipFile, err := os.Create(zipFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	// 解析 PDF，把图片写入 ZIP
	api.ExtractImages(pdfFile, nil, digestImage(zw), nil)
}

func digestImage(zw *zip.Writer) func(pdfcpu.Image, bool, int) error {
	return func(img pdfcpu.Image, singleImgPerPage bool, maxPageDigits int) error {
		myImg := (*MyImage)(unsafe.Pointer(&img))
		fn := fmt.Sprintf("%s-%d-%d.%s", myImg.Name, myImg.PageNr, myImg.ObjNr, myImg.FileType)
		f, err := zw.Create(fn)
		if err != nil {
			fmt.Println(err)
			return err
		}
		_, err = io.Copy(f, img.Reader)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println(fn)
		return nil
	}
}
