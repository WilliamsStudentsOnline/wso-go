package pictures

import "image"

func NewPictureBackendDummy() PictureBackend {
	return &pictureBackendDummy{}
}

type pictureBackendDummy struct{}

func (*pictureBackendDummy) SaveThumb(img image.Image, unixID string) error {
	return nil
}

func (*pictureBackendDummy) SaveLarge(img image.Image, unixID string) error {
	return nil
}

func (*pictureBackendDummy) Save(img image.Image, unixID string, category string) error {
	return nil
}
