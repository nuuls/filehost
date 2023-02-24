package diskstore

import (
	"io"
	"os"
	"path/filepath"

	"github.com/nuuls/filehost/internal/filestore"
	"github.com/pkg/errors"
)

type DiskStore struct {
	basePath string
}

var _ filestore.Filestore = &DiskStore{}

func New(basePath string) *DiskStore {
	return &DiskStore{
		basePath: basePath,
	}
}

func (d *DiskStore) Get(name string) (io.ReadSeeker, error) {
	file, err := os.Open(filepath.Join(d.basePath, name))
	if err != nil {
		return nil, errors.Wrap(err, "File not found")
	}
	return file, nil
}

func (d *DiskStore) Create(name string, data io.Reader) error {
	dstPath := filepath.Join(d.basePath, name)
	// TODO: check if file exists
	dst, err := os.Create(dstPath)
	if err != nil {
		return errors.Wrap(err, "Failed to create file on disk")
	}
	_, err = io.Copy(dst, data)
	if err != nil {
		return errors.Wrap(err, "Failed to write file to disk")
	}
	return nil
}

func (d *DiskStore) Delete(name string) error {
	err := os.Remove(filepath.Join(d.basePath, name))
	if err != nil {
		errors.Wrap(err, "Failed to delete file")
	}
	return nil
}
