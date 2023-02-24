package filestore

import "io"

type Filestore interface {
	Get(name string) (io.ReadSeeker, error)
	Create(name string, data io.Reader) error
	Delete(name string) error
}
