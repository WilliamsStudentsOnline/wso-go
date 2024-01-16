package pictures

import "image"

func NewPictureBackendDummy() PictureBackend {
	return &pictureBackendDummy{}
}

type pictureBackendDummy struct{}

func (d *pictureBackendDummy) DoesUserPhotoExists(unixID string) (bool, error) {
	return false, nil
}

func (*pictureBackendDummy) SaveUserPhotoThumb(unixID string, img image.Image) error {
	return nil
}

func (*pictureBackendDummy) SaveUserPhotoLarge(unixID string, img image.Image) error {
	return nil
}

func (*pictureBackendDummy) SaveEphmatchPhoto(unixID string, img image.Image) error {
	return nil
}

func (*pictureBackendDummy) DeleteEphmatchPhoto(unixID string) error {
	return nil
}

func (*pictureBackendDummy) SaveDormRoom(dormRoomID uint, reviewID uint, img image.Image) error {
	return nil
}

func (*pictureBackendDummy) ListDormRoom(dormRoomID uint) ([]string, error) {
	return []string{}, nil
}
