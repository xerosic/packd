package packd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/klauspost/compress/zstd"
)

type PIE [8]byte

type FirstHeader struct {
	isEncrypted   [1]byte
	miscellaneous [7]byte
}

type Header struct {
	pathLength uint16
	dataLength uint64
}

type Box struct {
	header Header
	path   []byte
	data   []byte
}

type PackdFile struct {
	PIE         PIE
	firstHeader FirstHeader
	boxes       []Box
}

func CreatePackdFileFromDirectory(path string, outputFileName string, isEncrypted bool) {
	// Create a new PackdFile
	pf := PackdFile{}

	// Set the PIE
	pf.PIE = PIE{0x3C, 0x21, 0x50, 0x41, 0x4B, 0x44, 0x21, 0x3E}

	boxes := []Box{}

	// Set the first header
	if isEncrypted {
		pf.firstHeader.isEncrypted = [1]byte{0x01}
	} else {
		pf.firstHeader.isEncrypted = [1]byte{0x00}
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
		header.pathLength = uint16(len(file.Name()))

		// Create a new box
		box := Box{}

		// Set the box header
		box.header = header

		// Set the box path
		box.path = []byte(file.Name())

		// Compress the data
		compressedData := encoder.EncodeAll(data, nil)

		if bytes.Equal(data, compressedData) {
			fmt.Printf("File %s was not compressed\n", file.Name())
		}

		// Set the box header data length
		header.dataLength = uint64(len(data))

		// Set the box data
		box.data = compressedData

		// Append the box to the boxes slice
		boxes = append(boxes, box)

	}

	// Set the PackdFile boxes
	pf.boxes = boxes

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
	if err := binary.Write(outFile, binary.LittleEndian, pf.firstHeader); err != nil {
		log.Fatal(err)
	}

	// Write each box
	for _, box := range pf.boxes {
		// Write header
		if err := binary.Write(outFile, binary.LittleEndian, box.header); err != nil {
			log.Fatal(err)
		}

		// Write path
		if _, err := outFile.Write(box.path); err != nil {
			log.Fatal(err)
		}

		// Write compressed data
		if _, err := outFile.Write(box.data); err != nil {
			log.Fatal(err)
		}
	}

	return
}
