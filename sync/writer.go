package sync

import (
	"io"
	"sync"

	"github.com/tidepool-org/platform/errors"
)

// CONCURRENCY: SAFE

func NewWriter(writer io.Writer) (*Writer, error) {
	if writer == nil {
		return nil, errors.New("writer is missing")
	}

	return &Writer{
		writer: writer,
	}, nil
}

type Writer struct {
	mutex  sync.Mutex
	writer io.Writer
}

func (w *Writer) Write(bytes []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.writer.Write(bytes)
}
