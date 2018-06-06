package main

import (
	"runtime"
	"flag"
	"fmt"
	"os"
	"log"
	"crypto/sha1"
	"io"
	"sync"
	"path/filepath"
)

type File = struct {
	FileName string
	Sha1 string
}

func makeSha1Sum(fileList chan string, out chan File) {
	var wg sync.WaitGroup

	sha1sum := func(path string) {
		//log.Printf("Reading file %v\n", path)
		f, err := os.Open(path)
		defer wg.Done()
		if err != nil {
			log.Printf("Error %s while reading %s\n", err, path)
			return
		}
		defer f.Close()
		h := sha1.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Printf("Can't make sha1 sum of %s\n", path)
			return
		}
		file := File{
			FileName: path,
			Sha1:     fmt.Sprintf("%x", h.Sum(nil)),
		}
		out <- file
	}
	i := 20
	for file := range fileList {
		wg.Add(1)
		go sha1sum(file)
		i--
		if i == 0 {
			// We use 20 threads to read files and make sha1sum
			i = 20
			wg.Wait()
		}
	}
	close(out)
}

func compare(out chan File, quit chan struct{}) {
	hashMap := make(map[string]string,100)
	for file := range out {
		if _, ok := hashMap[file.Sha1]; ok {
			log.Printf("Duplicate found at %s of %s, sha1sum %s\n",
				hashMap[file.Sha1], file.FileName, file.Sha1)
		} else {
			hashMap[file.Sha1] = file.FileName
		}
	}
	quit <- struct{}{}
}

func walker(filesList chan string, searchPath string, minSizeKb int64) {
	err := filepath.Walk(searchPath, func(path string, f os.FileInfo, err error) error {
		//Only append files which are not dirs and we don't need 2 skip that file
		if f == nil || f.IsDir() || ! f.Mode().IsRegular() {
			return nil
		}
		if minSizeKb != 0 && f.Size() < (minSizeKb*1024) {
			return nil
		}
		filesList <- path
		return nil
	})
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	close(filesList)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//Taken from https://gobyexample.com/command-line-flags
	var searchPath string
	flag.StringVar(&searchPath, "search_path", "./", "Directory, where we are going to search for our e-mails")
	var minSizeKb uint64
	flag.Uint64Var(&minSizeKb, "min_size_kb", 0, "Minimum value for file to be added to comparision (in KB)")

	flag.Parse()

	if searchPath == "" {
		fmt.Printf("Usage: %v -search_path=/some/path\n", os.Args[0])
		os.Exit(-1)
	}

	filesList := make(chan string)
	out := make(chan File)
	quit := make(chan struct{})
	go walker(filesList, searchPath, int64(minSizeKb))
	go makeSha1Sum(filesList, out)
	go compare(out, quit)
	<-quit
}
