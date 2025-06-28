package librarian

import (
	"librarian/logger"
)

type LibOptions struct {
	ConsoleOutput bool   // Prints the output - change this to an integer (lvls)
	UseHWAccel    bool   // this will use by default the vaapi thing for now
	FFmpegPath    string // Path to the ffmpeg binary
	FFprobePath   string // Path the ffprobe binary
	Format        string // Table or json available
	DBStore       bool   // Commit data to a database
	DryRun        bool   // No action is taken, only prints actions
}

func NewLibOptions() *LibOptions {
	return &LibOptions{
		ConsoleOutput: true,
		UseHWAccel:    false,
		FFmpegPath:    "/usr/bin/ffmpeg",
		FFprobePath:   "/usr/bin/ffprobe",
		Format:        "table",
	}
}

type Librarian struct {
	Options *LibOptions
}

func NewLibrarian(opts *LibOptions) (*Librarian, error) {
	// Check tooling
	if exists, err := pathExists(opts.FFmpegPath); !exists {
		return nil, err
	}

	if exists, err := pathExists(opts.FFprobePath); !exists {
		return nil, err
	}

	if !opts.ConsoleOutput {
		logger.SetOptions(opts.ConsoleOutput)
	}

	return &Librarian{
		Options: opts,
	}, nil
}

func (l *Librarian) ValidateFiles(sources []string) {
	cart := NewLibCart(sources)
	cart.validateFiles(l.Options)
}

func (l *Librarian) MoveFiles(sources []string, dest string) {
	cart := NewLibCart(sources)
	cart.CopyFiles(dest, l.Options)
}
