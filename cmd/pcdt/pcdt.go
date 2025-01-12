package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xerosic/packd"
)

func main() {
	// Define command line flags
	extract := flag.Bool("x", false, "Decompress a .pakd file")
	inputDir := flag.String("i", "", "Input directory to compress")
	inputFile := flag.String("f", "", "Input .pakd file to decompress")
	outputDir := flag.String("d", "", "Output directory for decompression")
	outputFile := flag.String("o", "output.pakd", "Output .pakd file")
	encrypt := flag.Bool("e", false, "Enable encryption")

	// Parse command line flags
	flag.Parse()

	// Check the mode
	if *extract {
		// Decompression mode
		if *inputFile == "" || *outputDir == "" {
			fmt.Println("Error: Input file and output directory are required for decompression")
			fmt.Println("Usage: pcdt -mode decompress -f <input.pakd> -d <output_dir>")
			flag.PrintDefaults()
			os.Exit(1)
		}

		// Check if input file exists
		if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
			fmt.Printf("Error: File '%s' does not exist\n", *inputFile)
			os.Exit(1)
		}

		// Extract the packd file
		packd.ExtractPackdFileToDirectory(*inputFile, *outputDir)
		fmt.Printf("Successfully extracted archive to: %s\n", *outputDir)
	} else {
		// Compression mode
		if *inputDir == "" {
			fmt.Println("Error: Input directory is required for compression")
			fmt.Println("Usage: pcdt -mode compress -i <input_dir> [-o output.pakd] [-e]")
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
}
