package librarian

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func getDurationVideo(path string) (float32, error) {
	if exists, err := ffprobeExists(); !exists {
		return 0, err
	}

	command := []string{
		"/usr/bin/ffprobe",
		"-v", "error",
		"-show_entries",
		"format=duration",
		"-of",
		"default=noprint_wrappers=1:nokey=1",
		path,
	}
	cmd := exec.Command(command[0], command[1:]...)
	fmt.Println(strings.Join(command, " "))

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

func VerifyMediaIntegrity(path string) (string, error) {
	if exists, err := ffmpegExists(); !exists {
		return "", err
	}

	duration, err := getDurationVideo(path)
	if err != nil {
		return "", err
	}
	// "-init_hw_device vaapi=va:/dev/dri/renderD128,driver=iHD -hwaccel vaapi -hwaccel_output_format"
	command := []string{
		"/usr/bin/ffmpeg",
		"-v", "error",
		"-init_hw_device", "vaapi=va:/dev/dri/renderD128,driver=iHD",
		"-hwaccel", "vaapi",
		"-hwaccel_output_format", "vaapi",
		"-i", path,
		"-f", "null", "-",
		"-progress", "pipe:1",
	}
	fmt.Println(strings.Join(command, " "))
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
					fmt.Println("Error reading stderr:", err)
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

			fmt.Printf("\rProgress: %.2f", percentage)
		} else if strings.HasPrefix(line, "progress=end") {
			fmt.Println("\nDone!")
		}
	}

	err = cmd.Wait()
	if killErr != nil {
		fmt.Println()
		return "", fmt.Errorf("aborted due to error output: %w", killErr)
	}

	if err != nil {
		fmt.Println()
		return "", fmt.Errorf("ffmpeg exited with error: %w", err)
	}

	return "", nil
}
