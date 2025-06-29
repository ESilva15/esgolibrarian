package librarian

import (
	"bufio"
	"fmt"
	"librarian/logger"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func getDurationVideo(path string, opts *LibOptions) (float32, error) {
	command := []string{
		opts.FFprobePath,
		"-v", "error",
		"-show_entries",
		"format=duration",
		"-of",
		"default=noprint_wrappers=1:nokey=1",
		path,
	}
	logger.Println(logrus.InfoLevel, strings.Join(command, " "))
	cmd := exec.Command(command[0], command[1:]...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start ffprobe: %w", err)
	}

	reader := bufio.NewReader(stdoutPipe)
	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(line), 32)
	if err != nil {
		return 0, err
	}

	return float32(duration), nil
}
