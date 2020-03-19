package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/tealeg/xlsx"
)

type StringSlice []string

func (i *StringSlice) String() string {
	return fmt.Sprintf("%v", []string(*i))
}
func (i *StringSlice) Set(s string) error {
	*i = append(*i, s)
	return nil
}

func main() {
	var (
		ImportPaths     StringSlice
		ImportFiles     StringSlice
		ProtoOutputPath string
		TableOutputPath string
		PackageName     string
	)
	flag.Var(&ImportPaths, "I", "import path(s)")
	flag.Var(&ImportFiles, "i", "import file(s)")
	flag.StringVar(&ProtoOutputPath, "P", "", "[REQUIRED] proto output path")
	flag.StringVar(&TableOutputPath, "O", "", "[REQUIRED] tbl output path")
	flag.StringVar(&PackageName, "p", "", "[REQUIRED] proto package")
	flag.Parse()
	if flag.NArg() == 0 || ProtoOutputPath == "" ||
		TableOutputPath == "" || PackageName == "" {
		flag.Usage()
		os.Exit(1)
	}
	// if a directory is spicified, traverse it
	filePaths := make([]string, 0)
	if f, err := os.Stat(flag.Arg(0)); err != nil {
		log.Fatal(err)
	} else if f.IsDir() {
		filepath.Walk(
			flag.Arg(0),
			func(pathPath string, f os.FileInfo, err error) (_ error) {
				if f == nil || f.IsDir() {
					return
				}
				filePaths = append(filePaths, pathPath)
				return
			})
	} else {
		filePaths = append(filePaths, flag.Arg(0))
	}
	var wg sync.WaitGroup
	for _, filePath := range filePaths {
		file, err := xlsx.OpenFile(filePath)
		if err != nil {
			log.Fatalf("open file %q fail: %v", filePath, err)
			return
		}
		for _, sheet := range file.Sheets {
			// ignore sheet prefixed by '#'
			if sheet.Name == "" || sheet.Name[0] == '#' {
				continue
			}
			x := &Maker{
				FilePath:        filePath,
				PackageName:     PackageName,
				ImportPaths:     ImportPaths,
				ImportFiles:     ImportFiles,
				ProtoOutputPath: ProtoOutputPath,
				TableOutputPath: TableOutputPath,
				Sheet:           sheet,
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := x.Make(); err != nil {
					fullname := UpperCamel(x.FileName, x.Sheet.Name)
					log.Printf("handle sheet %q fail: %v", fullname, err)
				}
			}()
		}
	}
	wg.Wait()
	return
}
