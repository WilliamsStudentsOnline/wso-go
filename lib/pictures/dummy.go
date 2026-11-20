package pictures

import (
	"image"
	"os"
)

func NewPictureBackendDummy() PictureBackend {
	return &pictureBackendDummy{}
}

type pictureBackendDummy struct{}

func (d *pictureBackendDummy) DoesUserPhotoExists(unixID string) (bool, error) {
	return false, nil
}

func (d *pictureBackendDummy) SaveUserPhotoBoth(unixID string, img image.Image) error {
	return nil
}

func (*pictureBackendDummy) SaveUserPhotoThumb(unixID string, img image.Image) error {
	return nil
}

func (*pictureBackendDummy) SaveUserPhotoLarge(unixID string, img image.Image) error {
	return nil
}

func (*pictureBackendDummy) SaveDormRoom(dormRoomID uint, reviewID uint, img image.Image) error {
	return nil
}

func (*pictureBackendDummy) ListDormRoom(dormRoomID uint) ([]string, error) {
	return []string{}, nil
}

func (*pictureBackendDummy) OpenPicture(relPath string) (*os.File, error) {
	return nil, os.ErrNotExist
}
