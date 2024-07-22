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
	dir     string
	name    string
	search  string
	workers int
	version string
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

func runFind(cmd *cobra.Command, args []string) {
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
