package service

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

type ImageStore interface {
	Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error) // returns image id and error
}

type DiskImageStore struct {
	mutex sync.RWMutex
	ImageFolder string
	Images map[string]*ImageInfo
}

type ImageInfo struct {
	LaptopID string
	Type string
	Path string
}

func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore {
		ImageFolder: imageFolder,
		Images: make(map[string]*ImageInfo),
	}
}

func (store *DiskImageStore) Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error) {
	// create new image ID
	imageId, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot create new image ID: %v", err)
	}

	imagePath := fmt.Sprintf("%s/%s.%s", store.ImageFolder, imageId.String(), imageType)

	file, err := os.Create(imagePath) // create new file
	if err != nil {
		return "", fmt.Errorf("cannot create the image file: %v", err)
	}

	_, err = imageData.WriteTo(file) // write image data to file that created before
	if err != nil {
		return "", fmt.Errorf("cannot write to image file: %v", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.Images[imageId.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type: imageType,
		Path: imagePath,
	}

	return imageId.String(), nil
} 


