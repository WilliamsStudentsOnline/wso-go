package local

import (
	"errors"
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
	maxDormtrakPhotos   = 10
)

var ErrorMaxDormtrakPhotos = errors.New("max dormtrak photos")

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
	log  *zap.SugaredLogger
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

	return &Backend{path: absPath, log: log}, nil
}

// Simply replace user profile thumb here
// All file names should be in the format `user/thumb/{userID}.jpg`
func (b *Backend) SaveUserPhotoThumb(unixID string, img image.Image) error {
	return b.saveUserProfile(img, unixID, dirUserThumb)
}

// Simply replace user profile large here
// All file names should be in the format `user/large/{userID}.jpg`
func (b *Backend) SaveUserPhotoLarge(unixID string, img image.Image) error {
	return b.saveUserProfile(img, unixID, dirUserLarge)
}

func (b *Backend) saveUserProfile(img image.Image, unixID string, category string) error {
	path := filepath.Join(b.path, category, unixID+".jpg")

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return err
	}

	return file.Close()
}

// All file names should be in the format `dormtrak/dormroom/{dormRoomID}/{reviewID}_{n}.jpg`
func (b *Backend) SaveDormRoom(dormRoomID uint, reviewID uint, img image.Image) error {
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

	// All file names should be in the format `dormRoomID/reviewID_n.jpg`
	// We parse find n+1
	maxN := -1
	for _, file := range files {
		var fileN, parsedReviewID int
		_, err := fmt.Sscanf(file.Name(), "%d_%d.jpg", &parsedReviewID, &fileN)
		if err != nil {
			continue
		}

		if uint(parsedReviewID) != reviewID {
			continue
		}

		if fileN > maxN {
			maxN = fileN
		}
	}

	if maxN > maxDormtrakPhotos {
		return ErrorMaxDormtrakPhotos
	}

	path := filepath.Join(dirPath, fmt.Sprintf("%d_%d.jpg", reviewID, maxN+1))

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return err
	}

	return file.Close()
}

func (b *Backend) ListDormRoom(dormRoomID uint) ([]string, error) {
	dirPath := filepath.Join(b.path, dirDormtrakDormroom, strconv.Itoa(int(dormRoomID)))

	if _, statErr := os.Stat(dirPath); os.IsNotExist(statErr) {
		return nil, statErr
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	pics := make([]string, len(files))

	for i := range files {
		pics[i] = files[i].Name()
	}

	return pics, nil
}

// All file names should be in the format `ephmatch/{userID}.jpg`
func (b *Backend) SaveEphmatchPhoto(unixID string, img image.Image) error {
	return b.saveUserProfile(img, unixID, dirEphmatch)
}

func (b *Backend) DeleteEphmatchPhoto(unixID string) error {
	fPath := filepath.Join(b.path, dirEphmatch, unixID+".jpg")

	if _, statErr := os.Stat(fPath); os.IsNotExist(statErr) {
		return statErr
	}

	return os.Remove(fPath)
}
