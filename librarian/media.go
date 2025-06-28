package librarian

import (
	"errors"
	"librarian/logger"
	"path/filepath"
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
	if val, err := pathExists(dest); !val {
		return err.Error(), ErrDestPathNotExist
	}

	if !isDir(dest) {
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
	errDetails, err := VerifyMediaIntegrity(m.Path, opts)
	m.AnalysisDuration = time.Now().UTC().Sub(analysisStart) / 1e6

	if err != nil {
		m.FailMediaIntegrity(errDetails, err)
	}
}

func (m *Media) CopyFile(opts *LibOptions) {
	analysisStart := time.Now()
	err := CopyFile(m.Path, m.DestPath)
	m.CopyDuration = time.Now().UTC().Sub(analysisStart) / 1e6

	if err != nil {
		m.FailMediaCopy(err)
	}
}
