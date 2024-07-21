package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/sync/errgroup"
)

var (
	dir            string
	name           string
	search         string
	workers        int
	version        string
	displayVersion bool
)

func main() {
	flag.StringVar(&dir, "dir", ".", "directory to search")
	flag.StringVar(&name, "name", "*.txt", "file name pattern to match")
	flag.StringVar(&search, "search", "", "string to search for")
	flag.IntVar(&workers, "c", 4, "maximum concurrency to use for file searching")
	flag.BoolVar(&displayVersion, "version", false, "display the version number")
	flag.Parse()

	if displayVersion {
		fmt.Println(version)
		return
	}

	if search == "" {
		flag.Usage()
		os.Exit(2)
	}

	fileChan := make(chan string, 100)

	go func() {
		if err := getMatchingFiles(dir, name, fileChan); err != nil {
			close(fileChan)
			log.Fatalf("error finding files: %v", err)
		}
		close(fileChan)
	}()

	if err := searchFiles(fileChan, search); err != nil {
		log.Fatalf("error searching files: %v", err)
	}
}

func getMatchingFiles(root, pattern string, fileChan chan<- string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		matched, err := filepath.Match(pattern, d.Name())
		if err != nil {
			return err
		}

		if matched && !d.IsDir() {
			fileChan <- path
		}
		return nil
	})
}

func searchFiles(fileChan <-chan string, searchStr string) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)

	eg, ctx := errgroup.WithContext(context.Background())

	for file := range fileChan {
		file := file
		sem <- struct{}{}
		wg.Add(1)
		eg.Go(func() error {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := searchInFile(file, searchStr); err != nil {
					return err
				}
			}
			<-sem
			return nil
		})
	}

	wg.Wait()
	close(sem)

	return eg.Wait()
}

func searchInFile(file, searchStr string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.EACCES) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), searchStr) {
			fmt.Println(file)
			break
		}
	}

	return scanner.Err()
}
