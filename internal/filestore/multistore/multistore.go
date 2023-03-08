package multistore

import (
	"io"

	"github.com/nuuls/filehost/internal/filestore"
)

var _ filestore.Filestore = &Multistore{}

type Multistore struct {
	stores []filestore.Filestore
}

// New creates a new Multistore.
// Preferred store should be first in the list.
func New(stores []filestore.Filestore) *Multistore {
	return &Multistore{
		stores: stores,
	}
}

// Get file from first available store
func (m *Multistore) Get(name string) (io.ReadSeekCloser, error) {
	var err error
	for _, s := range m.stores {
		var file io.ReadSeekCloser
		file, err = s.Get(name)
		if err == nil {
			return file, nil
		}
	}
	return nil, err
}

// Create file on first available store
func (m *Multistore) Create(name string, data io.Reader) error {
	var err error
	for _, s := range m.stores {
		err = s.Create(name, data)
		if err == nil {
			return nil
		}
	}
	return err
}

// Delete file from all stores, returns no error if deleted from any
func (m *Multistore) Delete(name string) error {
	var err error
	for _, s := range m.stores {
		err2 := s.Delete(name)
		if err != nil {
			err = err2
		}
	}
	return err
}
