package librarian

import (
	"encoding/json"
	"fmt"
	"librarian/logger"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type Cart struct {
	List map[string]*Media `json:"media"`
}

func NewLibCart(initialLoad []string) *Cart {
	newCart := &Cart{List: make(map[string]*Media)}
	for _, media := range initialLoad {
		newCart.AddMedia(media)
	}
	return newCart
}

func (c *Cart) AddMedia(path string) {
	key := path
	newMedia := NewMediaState()

	realpath, err := filepath.Abs(path)
	if err != nil {
		key = path
		newMedia.Err = ErrFailedToResolveRealpath
		newMedia.Details = err.Error()
		newMedia.State = false
	} else {
		key = realpath
	}

	c.List[key] = NewMediaState()
	c.List[key].Path = key
}

func (c *Cart) GetMediaState(path string) *Media {
	return c.List[path]
}

func (c *Cart) PrintSummary(opts *LibOptions) {
	// Log it, which is more useful than having it here truth be told
	logger.PrintJson(
		logrus.InfoLevel,
		func() map[string]any {
			result := make(map[string]any)
			summary := make(map[string]any, len(c.List))
			for k, v := range c.List {
				summary[k] = v
			}
			result["data"] = summary
			return result
		}(),
		"summary",
	)

	if opts.Format == "table" {
		fmt.Println("Summary:")
		for _, src := range c.List {
			err := ""
			if src.Err != nil {
				err = src.Err.Error()
			}
			fmt.Printf("%-6t %-40s %-33.33s %-33.33s %s\n",
				src.State,
				filepath.Base(src.Path),
				src.Hash,
				src.CopyHash,
				err,
			)
		}
	} else if opts.Format == "json" {
		jsonBytes, err := json.Marshal(c.List)
		if err != nil {
			logger.Printf(logrus.FatalLevel, "Failed to marshal cart data: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
	}
}

func (c *Cart) CopyFiles(dest string, options *LibOptions) {
	logger.Printf(logrus.InfoLevel, "Starting cart copying routine")

	// Iterate the given sources:
	// - Verify the integrity with ffmpeg
	// - Create the checksum of the original file
	// - Copy the original file to its destinty
	// - Create the checksum of the copied file
	// - Validate the copied file by its checksum
	for _, src := range c.List {
		// Add the new media to the cart (keep its state)
		curMedia := c.GetMediaState(src.Path)

		curMedia.ResolveDestpath(dest)
		if curMedia.Err != nil {
			continue
		}

		// Verify the integrity with ffmpeg
		curMedia.VerifyIntegrity(options)
		if curMedia.Err != nil {
			continue
		}

		// Store the checksum of the original file
		curMedia.ChecksumOriginal()
		if curMedia.Err != nil {
			continue
		}

		// Copy the source to its destiny
		curMedia.CopyFile(options)
		if curMedia.Err != nil {
			continue
		}

		// Get the checksum of the copy
		curMedia.ChecksumCopy()
		if curMedia.Err != nil {
			continue
		}

		// Validate the copy of the original file
		curMedia.ValidateChecksums()
		if curMedia.Err != nil {
			continue
		}
	}

	c.PrintSummary(options)
}

func (c *Cart) validateFiles(options *LibOptions) {
	for _, media := range c.List {
		media.VerifyIntegrity(options)
	}

	c.PrintSummary(options)
}
