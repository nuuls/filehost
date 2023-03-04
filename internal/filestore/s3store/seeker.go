package s3store

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3ReadSeeker struct {
	getRange func(string) (*s3.GetObjectOutput, error)

	buffer *bytes.Buffer

	offset int64
}

const bufferSize = 1024 * 1024 * 16 // 16MB

func newReadSeeker(getRange func(string) (*s3.GetObjectOutput, error)) *s3ReadSeeker {
	buf := make([]byte, 0, bufferSize)
	return &s3ReadSeeker{
		getRange: getRange,
		buffer:   bytes.NewBuffer(buf),
	}
}

func (r *s3ReadSeeker) Read(buf []byte) (int, error) {
	if r.buffer.Len() < len(buf) {
		rangeStr := fmt.Sprintf("bytes=%d-%d", r.offset, r.offset+bufferSize)
		res, err := r.getRange(rangeStr)
		if err != nil {
			return -1, err
		}
		defer res.Body.Close()

		r.buffer.Truncate(0)
		_, err = io.Copy(r.buffer, res.Body)
		if err != nil {
			return -1, err
		}
	}

	n, err := r.buffer.Read(buf)
	r.offset += int64(n)
	return n, err
}

func (r *s3ReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		r.offset += offset
		return r.offset, nil
	case io.SeekStart:
		r.offset = 0 + offset
		return r.offset, nil
	case io.SeekEnd:
		res, err := r.getRange("bytes=0-")
		if err != nil {
			return -1, err
		}
		defer res.Body.Close()
		r.offset = res.ContentLength + offset
		return r.offset, nil
	}

	return -1, fmt.Errorf("Invalid seek whence")
}
