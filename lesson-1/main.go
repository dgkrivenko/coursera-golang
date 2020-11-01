package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

// lineCreate - create line for printing
func lineCreate(file os.FileInfo, layer, lastIdx int, selfLast bool) string {
	line := strings.Repeat("│\t", lastIdx)
	line += strings.Repeat("\t", layer-lastIdx)

	if selfLast {
		line += "└───"
	} else {
		line += "├───"
	}
	if file.IsDir() {
		line += fmt.Sprintf("%v", file.Name())
	} else {
		if file.Size() == 0 {
			line += fmt.Sprintf("%v (empty)", file.Name())
		} else {
			line += fmt.Sprintf("%v (%vb)", file.Name(), file.Size())
		}
	}
	return line
}

// deleteFiles - delete files from slice
func deleteFiles(files []os.FileInfo) (result []os.FileInfo) {
	for _, f := range files {
		if f.IsDir() {
			result = append(result, f)
		}
	}
	return
}

// printDir - traverses directories and prints the result
func printDir(out io.Writer, path string, printFiles bool) error {
	rootFiles, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("error while read dir %v: %v", path, err)
	}

	if !printFiles {
		rootFiles = deleteFiles(rootFiles)
	}

	sort.SliceStable(rootFiles, func(i, j int) bool {
		return rootFiles[i].Name() > rootFiles[j].Name()
	})

	dirs := [][]os.FileInfo{rootFiles}
	// fullPath - contains directory names from root to current
	fullPath := []string{path}

	// verticalLayer - use for print vertical lines
	var verticalLayer int

DirsLoop:
	for len(dirs) > 0 {
		dir := &(dirs[len(dirs)-1])

		for len(*dir) > 0 {
			selfLast := len(*dir)-1 == 0
			f := (*dir)[len(*dir)-1]
			*dir = (*dir)[:len(*dir)-1]
			if f.IsDir() {
				fullPath = append(fullPath, f.Name())
				readPath := strings.Join(fullPath, "/")
				files, err := ioutil.ReadDir(readPath)
				if err != nil {
					return fmt.Errorf("error while read dir %v:%v", readPath, err)
				}
				if !printFiles {
					files = deleteFiles(files)
				}

				sort.SliceStable(files, func(i, j int) bool {
					return files[i].Name() > files[j].Name()
				})
				_, err = fmt.Fprintln(out, lineCreate(f, len(dirs)-1, verticalLayer, selfLast))
				if err != nil{
					return fmt.Errorf("error while print result: %v", err)
				}
				if !selfLast {
					verticalLayer += 1
				}
				dirs = append(dirs, files)
				continue DirsLoop
			} else {
				_, err = fmt.Fprintln(out, lineCreate(f, len(dirs)-1, verticalLayer, selfLast))
				if err != nil{
					return fmt.Errorf("error while print result: %v", err)
				}
			}
		}
		fullPath = fullPath[:len(fullPath)-1]
		dirs = dirs[:len(dirs)-1]
		if verticalLayer > len(dirs)-1 {
			verticalLayer = len(dirs) - 1
		}
	}
	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	sort.SliceStable(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})
	err = printDir(out, path, printFiles)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
