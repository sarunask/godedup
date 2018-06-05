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

func makeSha1Sum(fileList *[]string) <-chan File {
	var wg sync.WaitGroup
	out := make(chan File, len(*fileList))

	sha1sum := func(path string) {
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
	for _, file := range *fileList {
		wg.Add(1)
		go sha1sum(file)
		i--
		if i == 0 {
			i = 20
			wg.Wait()
		}
	}
	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//Taken from https://gobyexample.com/command-line-flags
	//1. Do this for the entire set of OBI mails until now?
	var searchPath string
	flag.StringVar(&searchPath, "search_path", "./", "Directory, where we are going to search for our e-mails")

	flag.Parse()

	if searchPath == "" {
		fmt.Printf("Usage: %v -search_path=/some/path\n", os.Args[0])
		os.Exit(-1)
	}

	filesList := make([]string, 0)
	//go testEmail(fileList)
	err := filepath.Walk(searchPath, func(path string, f os.FileInfo, err error) error {
		//Only append files which are not dirs and we don't need 2 skip that file
		if f != nil && f.IsDir() == false && f.Mode().IsRegular() == true {
			filesList = append(filesList, path)
		}
		return nil
	})
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	hashMap := make(map[string]string,100)
	for file := range makeSha1Sum(&filesList) {
		if _, ok := hashMap[file.Sha1]; ok {
			log.Printf("Duplicate found at %s of %s, sha1sum %s\n",
				hashMap[file.Sha1], file.FileName, file.Sha1)
		} else {
			hashMap[file.Sha1] = file.FileName
		}
	}
}
