package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knasan/parsegp"
)

var fileList []string

// isFile checks if the given path is a file.
//
// The function takes a string parameter `path` representing the file path to be checked.
// It uses the os.Stat function to retrieve information about the file at the given path.
// If an error occurs during the stat operation, the function returns false.
// Otherwise, it checks if the file is a directory (info.IsDir()) and returns the opposite boolean value.
// If the file is a directory, the function returns false; otherwise, it returns true.
func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// processFile opens the file at the given path, reads its header information,
// and prints the extracted details.
//
// Parameters:
// - path: A string representing the file path to be processed.
//
// Returns:
//   - An error if any error occurs during file opening, reading, or printing the header information.
//     Otherwise, it returns nil.
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
		"File: %s\n"+
			"GP-Version: %s\n"+
			"Artist: %s\n"+
			"Title: %s\n"+
			"Subtitle: %s\n"+
			"Album: %s\n"+
			"Lyric: %s\n"+
			"Music: %s\n"+
			"Copyright: %s\n"+
			"Transcriber/Tab: %s\n"+
			"Notice/Instruction: %s\n",
		gp.FullPath,
		gp.Version,
		gp.Artist,
		gp.Title,
		gp.Subtitle,
		gp.Album,
		gp.LyricsAuthor,
		gp.MusicAuthor,
		gp.Copyright,
		gp.Tab,
		gp.Instructions)
	fmt.Println("--")
	return nil
}

// run recursively walks the given path and collects all files within it.
// It returns a slice of file paths and an error if any occurs during the walk operation.
//
// Parameters:
//   - path: A string representing the root directory to start walking from.
//
// Returns:
//   - A slice of strings containing the file paths found during the walk.
//   - An error if any error occurs during the walk operation.
func run(path string) ([]string, error) {
	list, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, d := range list {
		if d.IsDir() {
			_, err := run(filepath.Join(path, d.Name()))
			if err != nil {
				return nil, err
			}
		} else {
			fileList = append(fileList, filepath.Join(path, d.Name()))
		}
	}

	return fileList, nil
}

// main is the entry point of the program. It walks the path and processes the files.
// If a file path is provided as a command-line argument, it processes that single file.
// Otherwise, it walks the current directory and processes all supported files found.
func main() {
	var err error
	// Check if a file path is provided as a command-line argument
	if len(os.Args) > 1 && isFile(os.Args[1]) {
		// Process the single file provided as a command-line argument
		if err = processFile(os.Args[1]); err != nil {
			fmt.Println(err)
		}
	} else {
		// Walk the current directory and collect all files
		if _, err := run("."); err != nil {
			panic(err)
		}

		// Process each collected file
		for _, file := range fileList {
			ext := filepath.Ext(file)
			// Check if the file extension is supported
			for _, format := range parsegp.SupportedFormats() {
				if ext == format {
					// Process the supported file
					if err = processFile(file); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}
