package librarian

import (
	"fmt"
)

type progressWriter struct {
	total      int64
	written    int64
	lastOutput int
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.written += int64(n)

	percent := int(float64(pw.written) * 100 / float64(pw.total))
	if percent != pw.lastOutput {
		fmt.Printf("\rProgress: %d%%", percent)
		pw.lastOutput = percent
	}

	return n, nil
}
