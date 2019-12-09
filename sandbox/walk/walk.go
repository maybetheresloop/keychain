package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	dir := os.Args[1]
	//dir, err := filepath.Abs(dir)
	//if err != nil {
	//	log.Fatalf("not a valid path")
	//}

	fmt.Printf("abs: %s\n", dir)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != dir {
			//fmt.Printf("Skipping directory: %s\n", path)
			return filepath.SkipDir
		}

		fmt.Printf("visited file: %s\n", path)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
