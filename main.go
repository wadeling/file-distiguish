package main

import (
	"flag"
	"fmt"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/essentialkaos/go-jar"
	"github.com/go-enry/go-enry/v2"
	ft "github.com/wadeling/file-distiguish/pkg/filetype"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	langCandidates = []string{"Python", "JavaScript", "Go", "Ruby", "PHP", "Shell", "Perl", "Jar"}
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

func detectJar(filename string) (bool, error) {
	manifest, err := jar.ReadFile(filename)
	if err != nil {
		return false, err
	}
	log.Printf("manifest.%+v", manifest)
	return true, nil
}

// detectLang accurate is bad
func detectLang(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	lexer := lexers.Analyse(string(data))
	if lexer == nil {
		return "", fmt.Errorf("failed to analyze file %v", filename)
	}
	return lexer.Config().Name, nil
}

func main() {
	fileName := flag.String("fileName", "", "file to detect")
	fileDirs := flag.String("dirs", "", "dir that contains language files for bench test")
	language := flag.String("lang", "", "language that dir refer to")
	checkJar := flag.Bool("checkJar", false, "check jar file")
	//fileExt := flag.String("ext", "", "language file ext")
	flag.Parse()

	// detect single file
	if len(*fileName) > 0 {
		if *checkJar {
			//isJar, err := detectJar(*jarFile)
			isJar, err := ft.IsJar(*fileName)
			log.Printf("is jar:%v,err:%v", isJar, err)
		} else {
			lang, safe := classifyFile(*fileName, langCandidates)
			//lang, err := detectLang(*fileName)
			log.Printf("lang %v,safe %v", lang, safe)
		}
		return
	}

	// bench with dir
	f, err := os.Open(*fileDirs)
	if err != nil {
		log.Fatalf("failed to open dir.%v,%v", *fileDirs, err)
		return
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
		if *checkJar {
			isJar, err := ft.IsJar(fullPath)
			//isJar, err := detectJar(fullPath)
			if err != nil {
				log.Printf("error: failed to check jar file.%v", file.Name())
				continue
			}
			if !isJar {
				log.Printf("not jar file.%v", file.Name())
				continue
			}
		} else {
			lang, err := classifyFile(fullPath, langCandidates)
			if err != nil {
				log.Printf("error: failed to class file.%v", file.Name())
				continue
			}
			if lang != *language {
				log.Printf("error: language not match.%v,%v", lang, *language)
				continue
			}
		}
		successNum += 1
	}
	end := time.Now()
	duration := end.Sub(start)
	log.Printf("total file num:%v,total size:%v,total time:%v,average time:%v,succ num:%v,accurate:%v,average file size:%v",
		totalNum, totalSize, duration, duration/time.Duration(totalNum), successNum, float32(successNum)/float32(totalNum), totalSize/totalNum)
}
