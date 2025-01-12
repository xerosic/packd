package packd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

type PIE [8]byte

type FirstHeader struct {
	IsEncrypted   [1]byte
	Miscellaneous [7]byte
}

type Header struct {
	PathLength uint16
	DataLength uint64
}

type Box struct {
	BoxHeader Header
	Path      []byte
	Data      []byte
}

type PackdFile struct {
	PIE         PIE
	FirstHeader FirstHeader
	Boxes       []Box
}

func CreatePackdFileFromDirectory(path string, outputFileName string, isEncrypted bool) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Create a new PackdFile
	pf := PackdFile{}

	// Set the PIE
	pf.PIE = PIE{0x3C, 0x21, 0x50, 0x41, 0x4B, 0x44, 0x21, 0x3E}

	boxes := []Box{}

	// Set the first header
	if isEncrypted {
		pf.FirstHeader.IsEncrypted = [1]byte{0x01}
	} else {
		pf.FirstHeader.IsEncrypted = [1]byte{0x00}
	}

	// Create zstd encoder
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		log.Fatal(err)
	}
	defer encoder.Close()

	// Iterate each file in the directory and add it to the PackdFile
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		// Open the file
		f, err := os.Open(path + "/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}

		// Read the file contents
		data, err := io.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()

		// Create a new header for the box
		header := Header{}

		// Set the box header path length
		header.PathLength = uint16(len(file.Name()))

		// Create a new box
		box := Box{}

		// Set the box header
		box.BoxHeader = header

		// Set the box path
		box.Path = []byte(file.Name())

		// Compress the data
		compressedData := encoder.EncodeAll(data, nil)

		if bytes.Equal(data, compressedData) {
			fmt.Printf("File %s was not compressed\n", file.Name())
		}

		// Set the box header data length
		header.DataLength = uint64(len(data))

		// Set the box data
		box.Data = compressedData

		// Append the box to the boxes slice
		boxes = append(boxes, box)

	}

	// Set the PackdFile boxes
	pf.Boxes = boxes

	// Create the output file
	outFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	// Write PIE
	if err := binary.Write(outFile, binary.LittleEndian, pf.PIE); err != nil {
		log.Fatal(err)
	}

	// Write FirstHeader
	if err := binary.Write(outFile, binary.LittleEndian, pf.FirstHeader); err != nil {
		log.Fatal(err)
	}

	// Write each box
	for _, box := range pf.Boxes {
		// Write header
		if err := binary.Write(outFile, binary.LittleEndian, box.BoxHeader); err != nil {
			log.Fatal(err)
		}

		// Write path
		if _, err := outFile.Write(box.Path); err != nil {
			log.Fatal(err)
		}

		// Write compressed data
		if _, err := outFile.Write(box.Data); err != nil {
			log.Fatal(err)
		}
	}

	return
}

func ExtractPackdFileToDirectory(inputFileName string, outputPath string) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Open input file
	inFile, err := os.Open(inputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	// Create decoder
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer decoder.Close()

	// Read and verify PIE
	pie := PIE{}
	if err := binary.Read(inFile, binary.LittleEndian, &pie); err != nil {
		log.Fatal(err)
	}
	expectedPIE := PIE{0x3C, 0x21, 0x50, 0x41, 0x4B, 0x44, 0x21, 0x3E}
	if pie != expectedPIE {
		log.Fatal("error: invalid file format (no PIE present / invalid PIE)")
	}

	// Read FirstHeader
	firstHeader := FirstHeader{}
	if err := binary.Read(inFile, binary.LittleEndian, &firstHeader); err != nil {
		log.Fatal(err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		log.Fatal(err)
	}

	// Read and extract boxes
	for {
		// Read header
		header := Header{}
		if err := binary.Read(inFile, binary.LittleEndian, &header); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		// Read path
		path := make([]byte, header.PathLength)
		if _, err := io.ReadFull(inFile, path); err != nil {
			log.Fatal(err)
		}

		// Read compressed data
		compressedData := make([]byte, header.DataLength)
		if _, err := io.ReadFull(inFile, compressedData); err != nil {
			log.Fatal(err)
		}

		// Decompress data
		decompressedData, err := decoder.DecodeAll(compressedData, nil)
		if err != nil {
			log.Fatal(err)

			// Write to file
			outFilePath := filepath.Join(outputPath, string(path))
			if err := os.WriteFile(outFilePath, decompressedData, 0644); err != nil {
				log.Fatal(err)
			}
		}

	}
	return
}
