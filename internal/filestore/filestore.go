package filestore

import "io"

type Filestore interface {
	Get(name string) (io.ReadSeekCloser, error)
	Create(name string, data io.Reader) error
	Delete(name string) error
}
