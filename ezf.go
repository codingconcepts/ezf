package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	dir         string
	name        string
	search      string
	ignorePaths = []string{"node_modules"}
	workers     int
	version     string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ezf",
		Short: "Find files easily from the command line.",
		Run:   runFind,
	}

	rootCmd.Flags().StringVarP(&dir, "dir", "d", ".", "directory to search")
	rootCmd.Flags().StringVarP(&name, "name", "n", "", "file name pattern to match")
	rootCmd.Flags().StringVarP(&search, "search", "s", "", "string to search for")
	rootCmd.Flags().StringSliceVarP(&ignorePaths, "ignore", "i", nil, "a path to ignore (e.g. node_modules)")
	rootCmd.Flags().IntVarP(&workers, "concurrency", "c", 4, "maximum concurrency to use for file searching")
	rootCmd.MarkFlagRequired("search")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version of ezf.",
		Run:   runVersion,
	}

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Println(version)
}

type result struct {
	file string
	line string
}

func runFind(cmd *cobra.Command, args []string) {
	fileChan := make(chan string, 100)

	go func() {
		if err := getMatchingFiles(dir, name, fileChan); err != nil {
			close(fileChan)
			log.Fatalf("error finding files: %v", err)
		}
		close(fileChan)
	}()

	foundChan := make(chan result, 100)
	finished := make(chan struct{})
	go func() {
		for res := range foundChan {
			fmt.Println(res.file)
		}
		finished <- struct{}{}
	}()

	if err := searchFiles(fileChan, foundChan, search); err != nil {
		log.Fatalf("error searching files: %v", err)
	}

	close(foundChan)
	<-finished
}

func getMatchingFiles(root, pattern string, fileChan chan<- string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored files.
		for _, ignore := range ignorePaths {
			if strings.Contains(path, ignore) {
				return nil
			}
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

func searchFiles(fileChan <-chan string, foundChan chan<- result, searchStr string) error {
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
				if err := searchInFile(file, searchStr, foundChan); err != nil {
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

func searchInFile(file, searchStr string, foundChan chan<- result) (err error) {
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
			// fmt.Println(file)
			foundChan <- result{
				file: file,
				line: scanner.Text(),
			}
			break
		}
	}

	return scanner.Err()
}
