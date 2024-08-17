package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knasan/parsegp"
)

var fileList []string

// isFile checks if the path is a file
func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// processFile opens the file and prints the header information
func processFile(path string) error {

	gp, err := parsegp.NewGPFile(path)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", path, err)
	}
	fmt.Println("--")

	if err := gp.LoadHeader(); err != nil {
		return fmt.Errorf("error reading file %s: %v", path, err)
	}

	fmt.Printf(
		"File: %s"+
			"\nGP-Version: %s"+
			"\nArtist: %s"+
			"\nTitle: %s"+
			"\nSubtitle: %s"+
			"\nAlbum: %s"+
			"\nLyric: %s"+
			"\nMusic: %s"+
			"\nCopyright: %s"+
			"\nTransciber: %s"+
			"\nNotice: %s\n",
		gp.FullPath,
		gp.Version,
		gp.Artist,
		gp.Title,
		gp.Subtitle,
		gp.Album,
		gp.LyricBy,
		gp.MusicBy,
		gp.Copyright,
		gp.Transcriber,
		gp.Notice)
	fmt.Println("--")
	return nil
}

// run fill the repl type
func run(path string) ([]string, error) {
	list, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, d := range list {
		if d.IsDir() {
			_, err := run(filepath.Join(path, d.Name()))
			if err != nil {
				panic(err)
			}

		} else {
			fileList = append(fileList, filepath.Join(path, d.Name()))
		}
	}

	return fileList, err
}

// main walks the path and processes the files
func main() {
	var err error
	if len(os.Args) > 1 && isFile(os.Args[1]) {
		if err = processFile(os.Args[1]); err != nil {
			fmt.Println(err)
		}
	} else {
		if _, err := run("."); err != nil {
			panic(err)
		}

		for _, file := range fileList {
			ext := filepath.Ext(file)
			for _, format := range parsegp.SupportedFormats() {
				if ext == format {
					if err = processFile(file); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}
