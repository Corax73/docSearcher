package main

import (
	"archive/zip"
	"doc_searcher/docsearcher"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

func main() {
	var files []string
	var results []*docsearcher.Book

	err := filepath.WalkDir("docs", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("read dir with err zip: %w", err)
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("read dir with err zip: %w", err)
		return
	}
	fmt.Println(files)

	for _, filePath := range files {
		r, err := zip.OpenReader(filePath)
		if err != nil {
			fmt.Printf("open with err zip: %w", err)
		}
		defer r.Close()

		for _, f := range r.File {
			if f.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(f.Name), ".fb2") {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				fmt.Printf("read err %s from file: %w", f.Name, err)
			}

			info, err := docsearcher.ReadFB2(rc)
			if err != nil {
				panic(err)
			}
			rc.Close()
			if err != nil {
				fmt.Printf("parsing err %s: %w", f.Name, err)
			}

			results = append(results, info)
		}
	}
	for _, book := range results {
		fmt.Printf("Title: %s\n", book.Metadata.BookTitle)
		for _, a := range book.Metadata.Authors {
			fmt.Printf("Author: %s %s\n", a.FirstName, a.LastName)
		}
		fmt.Printf("Genre: %s\n", strings.Join(book.Metadata.Genres, ", "))
		fmt.Printf("Lang: %s\n", book.Metadata.Lang)
		fmt.Printf("Text length: %d\n", len(book.Text))
	}
}
