package librarian

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(orig string, dest string) error {
	srcFile, err := os.Open(orig)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// get total size
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	totalSize := srcInfo.Size()

	// create destination file
	dstFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Create a progress reporter
	pw := &progressWriter{total: totalSize}
	writer := io.MultiWriter(dstFile, pw)

	// copy the file
	writtenBytes, err := io.Copy(writer, srcFile)
	if err != nil {
		return err
	}

	fmt.Printf("\nCopy complete. Copied %d bytes\n", writtenBytes)
	return nil
}

func MoveFiles(sources []string, dest string) {
	mediaCart := NewLibCart(sources)

	// Iterate the given sources:
	// - Verify the integrity with ffmpeg
	// - Create the checksum of the original file
	// - Copy the original file to its destinty
	// - Create the checksum of the copied file
	// - Validate the copied file by its checksum
	for _, src := range sources {
		// Add the new media to the cart (keep its state)
		curMedia := mediaCart.GetMediaState(src)

		// Set the destiny path
		curMedia.SetDestPath(dest)

		curMedia.destPath = filepath.Join(dest, filepath.Base(src))
		fmt.Printf("%-50s -> %s\n", curMedia.path, curMedia.destPath)

		// Verify the integrity with ffmpeg
		errDetails, err := VerifyMediaIntegrity(src)
		if err != nil {
			curMedia.FailMediaIntegrity(errDetails, err)
			continue
		}

		// Store the checksum of the original file
		curMedia.ChecksumOriginal()
		if curMedia.err != nil {
			continue
		}

		// Copy the source to its destiny
		CopyFile(curMedia.path, curMedia.destPath)

		// Get the checksum of the copy
		curMedia.ChecksumCopy()
		if curMedia.err != nil {
			continue
		}

		// Validate the copy of the original file
		curMedia.ValidateChecksums()
		if curMedia.err != nil {
			continue
		}
	}

	mediaCart.PrintSummary()
}

func ValidateFiles(sources []string) {
	cart := NewLibCart(sources)
	for _, media := range cart.originals {
		errDetails, err := VerifyMediaIntegrity(media.path)
		if err != nil {
			media.FailMediaIntegrity(errDetails, err)
			continue
		}
	}

	cart.PrintSummary()
}
