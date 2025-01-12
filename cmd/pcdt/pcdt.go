package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xerosic/packd"
)

func main() {
	// Define command line flags
	inputDir := flag.String("i", "", "Input directory to compress")
	outputFile := flag.String("o", "output.pakd", "Output .pakd file")
	encrypt := flag.Bool("e", false, "Enable encryption")

	// Parse command line flags
	flag.Parse()

	// Check if input directory is provided
	if *inputDir == "" {
		fmt.Println("Error: Input directory is required")
		fmt.Println("Usage: pcdt -i <input_dir> [-o output.pakd] [-e]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Check if input directory exists
	if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
		fmt.Printf("Error: Directory '%s' does not exist\n", *inputDir)
		os.Exit(1)
	}

	// Create the packd file
	packd.CreatePackdFileFromDirectory(*inputDir, *outputFile, *encrypt)
	fmt.Printf("Successfully created packd archive: %s\n", *outputFile)
}
