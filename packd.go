package packd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

func CreatePackdFileFromDirectory(path string, outputFileName string, isEncrypted bool) error {
	// Initialize output file
	outFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Write PIE (magic bytes)
	if _, err := outFile.Write([]byte{0x3C, 0x21, 0x50, 0x41, 0x4B, 0x44, 0x21, 0x3E}); err != nil {
		return err
	}

	// Write FirstHeader
	var firstHeader [8]byte
	if isEncrypted {
		firstHeader[0] = 0x01
	}
	if _, err := outFile.Write(firstHeader[:]); err != nil {
		return err
	}

	// Prepare Zstd encoder
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return err
	}
	defer encoder.Close()

	// Process directory files
	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		// Read file data
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Compress data
		compressed := encoder.EncodeAll(data, nil)

		// Create header
		header := make([]byte, 10)
		relPath, _ := filepath.Rel(path, filePath)
		binary.LittleEndian.PutUint16(header[0:2], uint16(len(relPath)))
		binary.LittleEndian.PutUint64(header[2:10], uint64(len(compressed)))

		// Write to archive
		if _, err := outFile.Write(header); err != nil {
			return err
		}
		if _, err := outFile.WriteString(relPath); err != nil {
			return err
		}
		if _, err := outFile.Write(compressed); err != nil {
			return err
		}

		return nil
	})
}

func ExtractPackdFileToDirectory(inputFileName string, outputPath string) error {
	// Open input file
	inFile, err := os.Open(inputFileName)
	if err != nil {
		return err
	}
	defer inFile.Close()

	// Verify PIE
	pie := make([]byte, 8)
	if _, err := io.ReadFull(inFile, pie); err != nil {
		return err
	}
	if !bytes.Equal(pie, []byte{0x3C, 0x21, 0x50, 0x41, 0x4B, 0x44, 0x21, 0x3E}) {
		return fmt.Errorf("invalid PIE signature")
	}

	// Read FirstHeader (skip encryption check for now)
	if _, err := io.ReadFull(inFile, make([]byte, 8)); err != nil {
		return err
	}

	// Prepare Zstd decoder
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return err
	}
	defer decoder.Close()

	// Process boxes
	headerBuf := make([]byte, 10)
	for {
		// Read header
		if _, err := io.ReadFull(inFile, headerBuf); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		pathLength := binary.LittleEndian.Uint16(headerBuf[0:2])
		dataLength := binary.LittleEndian.Uint64(headerBuf[2:10])

		// Read path
		path := make([]byte, pathLength)
		if _, err := io.ReadFull(inFile, path); err != nil {
			return fmt.Errorf("path read failed: %w", err)
		}

		// Read compressed data
		compressed := make([]byte, dataLength)
		if _, err := io.ReadFull(inFile, compressed); err != nil {
			return fmt.Errorf("data read failed: %w", err)
		}

		// Decompress data
		decompressed, err := decoder.DecodeAll(compressed, nil)
		if err != nil {
			return fmt.Errorf("decompression failed: %w", err)
		}

		// Write output file
		outPath := filepath.Join(outputPath, string(path))
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, decompressed, 0644); err != nil {
			return err
		}
	}

	return nil
}
