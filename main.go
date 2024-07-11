package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/essentialkaos/go-jar"
	"github.com/go-enry/go-enry/v2"
	ft "github.com/wadeling/file-distiguish/pkg/filetype"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	lang := enry.GetLanguage(filename, []byte(data))
	//lang, safe := enry.GetLanguageByContent(filename, []byte(data))

	//lang, safe := enry.GetLanguageByClassifier([]byte(data), candidates)
	//log.Printf("file %v,lang:%v,safe %v", filename, lang, safe)
	log.Printf("file %v,lang:%v,", filename, lang)
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

func shouldFilterFile(filename string) bool {
	// if image file.jpg etc..
	if enry.IsImage(filename) {
		log.Printf("image file.%v", filename)
		return true
	}

	// check if binary
	f, err := os.Open(filename)
	if err != nil {
		log.Printf("failed to open file %v", filename)
		return true
	}
	reader := bufio.NewReader(f)
	buf := make([]byte, 8000)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		log.Printf("failed to read file.%v", filename)
		return true
	}
	if n > 0 {
		if enry.IsBinary(buf[:n]) {
			log.Printf("binary file.%v", filename)
			return true
		}
	}

	return false
}

func detectLangByDir(rootPath string, checkJar bool, expectLang string) {
	var (
		totalSize  int64 = 0
		totalNum   int64 = 0
		successNum int64 = 0
	)

	start := time.Now()

	_ = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("failed to walk dir.%v,%v", path, err)
			return err
		}
		if d.IsDir() {
			log.Printf("%v is dir,continue", d.Name())
			return nil
		}
		info, err := d.Info()
		if err != nil {
			log.Printf("err: failed to get file info.%v", d.Name())
			return err
		}
		if shouldFilterFile(path) {
			//log.Printf("ignore file for it is image or binary.%v", path)
			return nil
		}

		totalSize = totalSize + info.Size()
		totalNum += 1
		if checkJar {
			// check jar file
			isJar, err := ft.IsJar(path)
			if err != nil {
				log.Printf("error: failed to check jar file.%v", path)
				return err
			}
			if !isJar {
				log.Printf("not jar file.%v", path)
				return nil
			}
		} else {
			// check other lang file
			lang, err := classifyFile(path, langCandidates)
			if err != nil {
				log.Printf("error: failed to class file.%v", path)
				// continue check other files,so return nil
				return nil
			}
			if strings.ToLower(lang) != strings.ToLower(expectLang) {
				log.Printf("error: language not match.%v,%v", lang, expectLang)
				return nil
			}
		}
		successNum += 1
		return nil
	})
	end := time.Now()
	duration := end.Sub(start)
	log.Printf("total file num:%v,total size:%v,total time:%v,average time:%v,succ num:%v,accurate:%v,average file size:%v",
		totalNum, totalSize, duration, duration/time.Duration(totalNum), successNum, float32(successNum)/float32(totalNum), totalSize/totalNum)
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
			// is normal file
			if enry.IsConfiguration(*fileName) {
				log.Printf("is conf file")
				return
			}

			lang, safe := classifyFile(*fileName, langCandidates)
			//lang, err := detectLang(*fileName)
			log.Printf("lang %v,safe %v", lang, safe)
		}
		return
	}

	// recursively detect dirs
	detectLangByDir(*fileDirs, *checkJar, *language)

	log.Print("end")
}
