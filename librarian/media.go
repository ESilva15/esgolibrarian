package librarian

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"librarian/logger"
	"librarian/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	ErrDestPathNotExist         = errors.New("destiny path doesn't exist")
	ErrDestNotDir               = errors.New("destiny path isn't a directory")
	ErrFailedOrigChecksum       = errors.New("failed checksum on source file")
	ErrFailedCopyChecksum       = errors.New("failed checksum on copied file")
	ErrFailedChecksumValidation = errors.New("failed checksum validation")
	ErrFailedMediaIntegrity     = errors.New("failed media integrity test")
	ErrFailedMediaCopy          = errors.New("failed to copy media")
	ErrFailedToResolveRealpath  = errors.New("failed to resolve realpath")
	ErrFailedToResolveDestpath  = errors.New("failed to resolve dest realpath")
)

type Media struct {
	FileSize         int64         `json:"file_size"`   // in bytes
	FileLength       float32       `json:"file_length"` // in milliseconds
	Path             string        `json:"path"`
	DestPath         string        `json:"dest_path"`
	Hash             string        `json:"hash"`
	CopyHash         string        `json:"copy_hash"`
	Err              error         `json:"err"`
	Details          string        `json:"details"`
	State            bool          `json:"state"`
	AnalysisDuration time.Duration `json:"analysis_duration"`
	CopyDuration     time.Duration `json:"copy_duration"`
}

func NewMediaState() *Media {
	return &Media{
		FileSize:         0,
		FileLength:       0,
		Path:             "",
		DestPath:         "",
		Hash:             "",
		CopyHash:         "",
		Err:              nil,
		Details:          "",
		State:            true,
		AnalysisDuration: 0,
		CopyDuration:     0,
	}
}

func validateDestPath(dest string) (string, error) {
	// Destiny path must be a directory (for now)
	if val, err := utils.PathExists(dest); !val {
		return err.Error(), ErrDestPathNotExist
	}

	if !utils.IsDir(dest) {
		return ErrDestNotDir.Error(), ErrDestNotDir
	}

	return "", nil
}

func (m *Media) FailMediaIntegrity(deets string, err error) {
	m.State = false
	m.Details = err.Error()
	m.Err = ErrFailedMediaIntegrity
}

func (m *Media) FailCheckSum(err error) {
	m.State = false
	m.Err = ErrFailedCopyChecksum
	m.Details = err.Error()
}

func (m *Media) FailCopyCheckSum(err error) {
	m.State = false
	m.Err = ErrFailedCopyChecksum
	m.Details = err.Error()
}

func (m *Media) UpdateOrigChecksum(checksum string) {
	m.Hash = checksum
}

func (m *Media) UpdateCopyChecksum(checksum string) {
	m.CopyHash = checksum
}

func (m *Media) FailChecksumValidation() {
	m.State = false
	m.Err = ErrFailedChecksumValidation
}

func (m *Media) FailMediaCopy(err error) {
	m.State = false
	m.Err = ErrFailedMediaCopy
	m.Details = err.Error()
}

func (m *Media) ChecksumOriginal() {
	originalChecksum, err := ChecksumFile(m.Path)
	if err != nil {
		err = ErrFailedOrigChecksum
		m.FailCheckSum(err)
	}
	m.UpdateOrigChecksum(originalChecksum)
}

func (m *Media) ChecksumCopy() {
	copyChecksum, err := ChecksumFile(m.DestPath)
	if err != nil {
		err = ErrFailedCopyChecksum
		m.FailCopyCheckSum(err)
	}
	m.UpdateCopyChecksum(copyChecksum)
}

func (m *Media) ResolveDestpath(dest string) {
	logger.Printf(logrus.InfoLevel, "Resolving dest path from %s\n", dest)

	// Set the destiny path
	errDetails, err := validateDestPath(dest)
	if err != nil {
		m.Err = err
		m.Details = errDetails
		m.State = false
		return
	}

	relativeDestpath := filepath.Join(dest, filepath.Base(m.Path))
	absDestpath, err := filepath.Abs(relativeDestpath)
	if err != nil {
		logger.Printf(logrus.ErrorLevel, "Failed to resolve dest path %v\n", err)
		m.Err = ErrFailedToResolveDestpath
		m.Details = err.Error()
		m.State = false
		return
	}

	m.DestPath = absDestpath
	logger.Printf(logrus.InfoLevel, "Destpath is %s\n", m.DestPath)
}

func (m *Media) ValidateChecksums() {
	if m.Hash != m.CopyHash {
		m.FailChecksumValidation()
	}
}

func (m *Media) VerifyIntegrity(opts *LibOptions) {
	analysisStart := time.Now()
	errDetails, err := m.IntegrityCheck(opts)
	m.AnalysisDuration = time.Now().UTC().Sub(analysisStart) / 1e6

	if err != nil {
		m.FailMediaIntegrity(errDetails, err)
	}
}

func (m *Media) SetFilesize() {
	file, err := os.Stat(m.Path)
	if err != nil {
		m.Err = err
		m.Details = err.Error()
		m.State = false
	}

	m.FileSize = int(file.Size())
}

func (m *Media) CopyFile(opts *LibOptions) {
	analysisStart := time.Now()
	err := CopyFile(m.Path, m.DestPath)
	m.CopyDuration = time.Now().UTC().Sub(analysisStart) / 1e6

	if err != nil {
		m.FailMediaCopy(err)
	}
}

func (m *Media) IntegrityCheck(options *LibOptions) (string, error) {
	var err error

	m.FileLength, err = getDurationVideo(m.Path, options)
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
		"-i", m.Path,
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
			percentage := (progressSec / m.FileLength) * 100

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
