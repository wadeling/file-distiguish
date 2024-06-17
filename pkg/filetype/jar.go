package filetype

import (
	"archive/zip"
	"fmt"
	"github.com/h2non/filetype"
	"log"
	"os"
	"strings"
)

// IsJar bad.some jar file without class file. eg. sprint-boot's resource-1.0.0.jar
func IsJar(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// 读取前 261 个字节以检测文件类型
	header := make([]byte, 261)
	if _, err := file.Read(header); err != nil {
		log.Printf("Failed to read file header: %v", err)
		return false, err
	}

	// 使用 filetype 库检测文件类型
	kind, _ := filetype.Match(header)
	if kind != filetype.Unknown {
		fmt.Printf("File type detected: %s\n", kind.Extension)
	} else {
		fmt.Println("Unknown file type")
	}

	// 检查文件是否为 JAR 文件
	if kind.Extension == "zip" {
		isJar, err := checkIfJar(filename)
		if err != nil {
			log.Printf("Failed to check if file is a JAR: %v", err)
			return false, err
		}
		if isJar {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, nil
}

func checkIfJar(filePath string) (bool, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return false, err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".class") {
			log.Printf("found java class file.%v", f.Name)
			return true, nil
		}
	}
	return false, nil
}
