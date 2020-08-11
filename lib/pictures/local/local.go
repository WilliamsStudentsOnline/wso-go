package local

import (
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"
)

const (
	dirUserThumb        = "user/thumb"
	dirUserLarge        = "user/large"
	dirEphmatch         = "ephmatch"
	dirDormtrakDormroom = "dormtrak/dormroom"
)

// Pictures backend for local filesystem images
/*
Telos Structure:
/user/large/unixid.jpg
/user/thumb/unixid.jpg
/dormtrak/dormroom/dorm_room_id/n.jpg
/ephmatch/unixid.jpg
*/

type Backend struct {
	path string
}

func NewBackend(path string, log *zap.SugaredLogger) (*Backend, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Ensure (some) directories exist
	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		err := os.MkdirAll(absPath, 0770)
		if err != nil {
			return nil, err
		}
	}

	// Warn if these sub directories don't exist
	requiredDirs := []string{dirUserThumb, dirUserLarge, dirEphmatch, dirDormtrakDormroom}
	for _, dir := range requiredDirs {
		fullPath := filepath.Join(absPath, dir)
		if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
			log.Warnf("missing picture directory: %s", fullPath)

			err := os.MkdirAll(fullPath, 0770)
			if err != nil {
				return nil, err
			}
		}
	}

	return &Backend{path: absPath}, nil
}

// Simply replace user profile thumb here
func (b *Backend) SaveUserPhotoThumb(unixID string, img image.Image) error {
	return b.saveUserProfile(img, unixID, dirUserThumb)
}

// Simply replace user profile large here
func (b *Backend) SaveUserPhotoLarge(unixID string, img image.Image) error {
	return b.saveUserProfile(img, unixID, dirUserLarge)
}

func (b *Backend) saveUserProfile(img image.Image, unixID string, category string) error {
	path := filepath.Join(b.path, category, unixID+".jpg")

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 80})
	return err
}

func (b *Backend) SaveDormRoom(dormRoomID uint, img image.Image) error {
	dirPath := filepath.Join(b.path, dirDormtrakDormroom, strconv.Itoa(int(dormRoomID)))

	if _, statErr := os.Stat(dirPath); os.IsNotExist(statErr) {
		err := os.MkdirAll(dirPath, 0770)
		if err != nil {
			return err
		}
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// All file names should be in the format `dormRoomID/n.jpg`
	// We parse find n+1
	maxN := -1
	for _, file := range files {
		var fileN int
		_, err := fmt.Sscanf(file.Name(), "%d.jpg", &fileN)
		if err != nil {
			continue
		}

		if fileN > maxN {
			fileN = maxN
		}
	}

	path := filepath.Join(dirPath, fmt.Sprintf("%d.jpg", maxN+1))

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 80})
	return err
}

func (b *Backend) SaveEphmatchPhoto(unixID string, img image.Image) error {
	return b.saveUserProfile(img, unixID, dirEphmatch)
}

func (b *Backend) DeleteEphmatchPhoto(unixID string) error {
	fPath := filepath.Join(b.path, dirDormtrakDormroom, unixID+".jpg")

	if _, statErr := os.Stat(fPath); os.IsNotExist(statErr) {
		return statErr
	}

	return os.Remove(fPath)
}
