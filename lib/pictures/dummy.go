package pictures

import "image"

func NewPictureBackendDummy() PictureBackend {
	return &pictureBackendDummy{}
}

type pictureBackendDummy struct{}

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

func (*pictureBackendDummy) SaveDormRoom(dormRoomID uint, img image.Image) error {
	return nil
}
