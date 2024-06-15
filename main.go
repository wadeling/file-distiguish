package main

import (
	"flag"
	"fmt"
	"github.com/go-enry/go-enry/v2"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	langCandidates = []string{"Python", "JavaScript", "Go", "Ruby", "PHP", "Shell", "Perl"}
)

func classifyFile(filename string, candidates []string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to read file:%v", err)
		return "", err
	}
	//	log.Printf("file data:%v", string(data))

	//lang := enry.GetLanguage(*fileName, []byte(data))
	//lang,safe := enry.GetLanguageByContent(*fileName, []byte(data))

	//candidates := []string{"Python", "JavaScript", "Go", "Ruby", "PHP", "Shell", "Perl"}
	lang, safe := enry.GetLanguageByClassifier([]byte(data), candidates)
	log.Printf("file %v,lang:%v,safe %v", filename, lang, safe)
	if len(lang) == 0 {
		log.Printf("failed to classify file type.%v", filename)
		return "", fmt.Errorf("failed to classify file type")
	}
	return lang, nil
}

func main() {
	fileName := flag.String("fileName", "", "file to detect")
	fileDirs := flag.String("dirs", "", "dir that contains language files for bench test")
	language := flag.String("lang", "", "language that dir refer to")
	//fileExt := flag.String("ext", "", "language file ext")
	flag.Parse()

	if len(*fileName) > 0 {
		data, err := os.ReadFile(*fileName)
		if err != nil {
			log.Fatalf("failed to read file:%v", err)
			return
		}
		log.Printf("file data:%v", string(data))

		//lang := enry.GetLanguage(*fileName, []byte(data))
		//lang,safe := enry.GetLanguageByContent(*fileName, []byte(data))

		candidates := []string{"Python", "JavaScript", "Go", "Ruby", "PHP", "Shell", "Perl"}
		lang, safe := enry.GetLanguageByClassifier([]byte(data), candidates)
		log.Printf("lang:%v,safe %v", lang, safe)
		return
	}

	// bench
	f, err := os.Open(*fileDirs)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// 读取目录下的所有文件和子目录
	files, err := f.Readdir(-1)
	if err != nil {
		log.Fatal(err)
		return
	}

	var (
		totalSize  int64 = 0
		totalNum   int64 = 0
		successNum int64 = 0
	)

	start := time.Now()
	for _, file := range files {
		if file.IsDir() {
			log.Printf("ignore subdir:%v", file.Name())
			continue
		}
		fullPath := filepath.Join(*fileDirs, file.Name())

		//ext := filepath.Ext(fullPath)
		//if *fileExt != ext {
		//	log.Printf("file %v not match,ignore", file.Name())
		//	continue
		//}

		totalSize = totalSize + file.Size()
		totalNum += 1
		lang, err := classifyFile(fullPath, langCandidates)
		if err != nil {
			log.Printf("error: failed to class file.%v", file.Name())
			continue
		}
		if lang != *language {
			log.Printf("error: language not match.%v,%v", lang, *language)
			continue
		}
		successNum += 1
	}
	end := time.Now()
	duration := end.Sub(start)
	log.Printf("total file num:%v,total size:%v,total time:%v,average time:%v,succ num:%v,accurate:%v,average file size:%v",
		totalNum, totalSize, duration, duration/time.Duration(totalNum), successNum, successNum/totalNum, totalSize/totalNum)
}
