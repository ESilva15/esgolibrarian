package librarian

import (
	"fmt"
	"path/filepath"
)

type Cart struct {
	originals map[string]*Media
}

func NewLibCart(initialLoad []string) *Cart {
	newCart := &Cart{originals: make(map[string]*Media)}
	for _, media := range initialLoad {
		newCart.AddMedia(media)
	}
	return newCart
}

func (c *Cart) AddMedia(path string) {
	c.originals[path] = NewMediaState()
	c.originals[path].path = path
}

func (c *Cart) GetMediaState(path string) *Media {
	return c.originals[path]
}

func (c *Cart) PrintSummary() {
	fmt.Println("Summary:")
	for _, src := range c.originals {
		err := ""
		if src.err != nil {
			err = src.err.Error()
		}
		fmt.Printf("%-6t %-40s %-33.33s %-33.33s %s\n",
			src.state,
			filepath.Base(src.path),
			src.hash,
			src.copyHash,
			err,
		)
	}
}
