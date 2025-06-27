package librarian

import (
	"errors"
)

var (
	ErrDestPathNotExist         = errors.New("destiny path doesn't exist")
	ErrDestNotDir               = errors.New("destiny path isn't a directory")
	ErrFailedOrigChecksum       = errors.New("failed checksum on source file")
	ErrFailedCopyChecksum       = errors.New("failed checksum on copied file")
	ErrFailedChecksumValidation = errors.New("failed checksum validation")
	ErrFailedMediaIntegrity     = errors.New("failed media integrity test")
)

type Media struct {
	path     string
	destPath string
	hash     string
	copyHash string
	copyPath string
	err      error
	details  string
	state    bool
}

func NewMediaState() *Media {
	return &Media{
		path:     "",
		destPath: "",
		hash:     "",
		err:      nil,
		details:  "",
		state:    true,
	}
}

func (m *Media) SetDestPath(dest string) {
	// Destiny path must be a directory (for now)
	if val, err := pathExists(dest); !val {
		m.state = false
		m.err = ErrDestPathNotExist
		m.details = err.Error()
		return
	}

	if !isDir(dest) {
		m.state = false
		m.err = ErrDestNotDir
		m.details = "Destiny path must be a directory"
	}
}

func (m *Media) FailMediaIntegrity(deets string, err error) {
	m.state = false
	m.details = deets
	m.err = ErrFailedMediaIntegrity
}

func (m *Media) FailCheckSum(err error) {
	m.state = false
	m.err = ErrFailedCopyChecksum
	m.details = err.Error()
}

func (m *Media) FailCopyCheckSum(err error) {
	m.state = false
	m.err = ErrFailedCopyChecksum
	m.details = err.Error()
}

func (m *Media) UpdateOrigChecksum(checksum string) {
	m.hash = checksum
}

func (m *Media) UpdateCopyChecksum(checksum string) {
	m.copyHash = checksum
}

func (m *Media) FailChecksumValidation() {
	m.state = false
	m.err = ErrFailedChecksumValidation
}

func (m *Media) ChecksumOriginal() {
	originalChecksum, err := ChecksumFile(m.path)
	if err != nil {
		err = ErrFailedOrigChecksum
		m.FailCheckSum(err)
	}
	m.UpdateOrigChecksum(originalChecksum)
}

func (m *Media) ChecksumCopy() {
	copyChecksum, err := ChecksumFile(m.destPath)
	if err != nil {
		err = ErrFailedCopyChecksum
		m.FailCopyCheckSum(err)
	}
	m.UpdateCopyChecksum(copyChecksum)
}

func (m *Media) ValidateChecksums() {
	if m.hash != m.copyHash {
		m.FailChecksumValidation()
	}
}
