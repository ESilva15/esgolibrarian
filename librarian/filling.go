package librarian

import (
	"io"
	"librarian/logger"
	"os"

	"github.com/sirupsen/logrus"
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

	logger.NoLogf("\n")
	logger.Printf(logrus.InfoLevel, "Copy complete. Copied %d bytes\n",
		writtenBytes)
	return nil
}
