package local

import (
	"errors"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

const (
	dirThumb = "thumb"
	dirLarge = "large"
)

// Pictures backend for local filesystem images

type Backend struct {
	path string
}

func NewBackend(path string, log *zap.SugaredLogger) (*Backend, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Ensure (some) directories exist
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, errors.New("missing picture directory: " + absPath)
	}

	// Warn if these sub directories don't exist
	if _, err := os.Stat(filepath.Join(absPath, dirThumb)); os.IsNotExist(err) {
		log.Warnf("missing picture sub-directory: %s", filepath.Join(absPath, dirThumb))
	}
	if _, err := os.Stat(filepath.Join(absPath, dirLarge)); os.IsNotExist(err) {
		log.Warnf("missing picture sub-directory: %s", filepath.Join(absPath, dirLarge))
	}

	return &Backend{path: absPath}, nil
}

func (b *Backend) SaveThumb(img image.Image, unixID string) error {
	return b.Save(img, unixID, dirThumb)
}

func (b *Backend) SaveLarge(img image.Image, unixID string) error {
	return b.Save(img, unixID, dirLarge)
}

func (b *Backend) Save(img image.Image, unixID string, category string) error {
	path := filepath.Join(b.path, category, unixID+".jpg")

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 80})
	return err
}
