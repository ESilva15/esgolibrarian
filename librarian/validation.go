package librarian

import (
	"bufio"
	"fmt"
	"io"
	"librarian/logger"
	"os/exec"
	"strconv"
	"strings"
	"sync"

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

func VerifyMediaIntegrity(path string, options *LibOptions) (string, error) {
	duration, err := getDurationVideo(path, options)
	if err != nil {
		return "", err
	}

	// "-init_hw_device vaapi=va:/dev/dri/renderD128,driver=iHD -hwaccel vaapi -hwaccel_output_format"
	hwaccelArgs := []string{
		"-init_hw_device", "vaapi=va:/dev/dri/renderD128,driver=iHD",
		"-hwaccel", "vaapi",
		"-hwaccel_output_format", "vaapi",
	}

	command := []string{options.FFmpegPath}
	if options.UseHWAccel {
		command = append(command, hwaccelArgs...)
	}
	command = append(command,
		"-v", "error",
		"-i", path,
		"-f", "null", "-",
		"-progress", "pipe:1",
	)
	logger.Println(logrus.InfoLevel, strings.Join(command, " "))
	cmd := exec.Command(command[0], command[1:]...)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var once sync.Once
	var killErr error

	go func() {
		reader := bufio.NewReader(stderrPipe)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					logger.Printf(logrus.ErrorLevel, "Error reading stderr: %v\n", err)
				}
				break
			}

			if strings.TrimSpace(line) != "" {
				once.Do(func() {
					killErr = cmd.Process.Kill()
				})
				break
			}
		}
	}()

	scanner := bufio.NewScanner(stdoutPipe)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "out_time_ms=") {
			progressStr := strings.TrimPrefix(line, "out_time_ms=")
			progressParsed, err := strconv.ParseFloat(progressStr, 32)
			if err != nil {
				return "", err
			}
			progressSec := float32(progressParsed) / 1000.0 / 1000.0
			percentage := (progressSec / duration) * 100

			logger.NoLogf("\rAnalyzing: %.2f%%", percentage)
		} else if strings.HasPrefix(line, "progress=end") {
			logger.NoLogf("\n")
			logger.Println(logrus.InfoLevel, "Done")
		}
	}

	err = cmd.Wait()
	if killErr != nil {
		logger.NoLogf("\n")
		return "", fmt.Errorf("aborted due to error output: %w", killErr)
	}

	if err != nil {
		logger.NoLogf("\n")
		return "", fmt.Errorf("ffmpeg exited with error: %w", err)
	}

	return "", nil
}
