package compress

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/klauspost/compress/zstd"

	"github.com/tidepool-org/platform/errors"
)

const CompressionLevel = zstd.SpeedDefault

func CompressReadCloser(reader io.Reader) *CompressedReadCloser {
	return &CompressedReadCloser{reader: reader}
}

type CompressedReadCloser struct {
	reader     io.Reader
	pipeReader *io.PipeReader
	pipeWriter *io.PipeWriter
}

func (l *CompressedReadCloser) Read(p []byte) (int, error) {
	if l.pipeReader == nil || l.pipeWriter == nil {
		if l.reader == nil {
			return 0, errors.New("reader is missing")
		}

		l.pipeReader, l.pipeWriter = io.Pipe()
		go func() {
			if encoder, err := zstd.NewWriter(l.pipeWriter, zstd.WithEncoderLevel(CompressionLevel)); err != nil {
				_ = l.pipeWriter.CloseWithError(err)
			} else if _, err = io.Copy(encoder, l.reader); err != nil {
				_ = encoder.Close()
				_ = l.pipeWriter.CloseWithError(err)
			} else if err = encoder.Close(); err != nil {
				_ = l.pipeWriter.CloseWithError(err)
			} else {
				_ = l.pipeWriter.Close()
			}
		}()
	}
	return l.pipeReader.Read(p)
}

func (l *CompressedReadCloser) Close() error {
	if l.pipeReader != nil {
		if err := l.pipeReader.Close(); err != nil {
			return err
		}
		l.pipeReader = nil
	}
	if l.pipeWriter != nil {
		if err := l.pipeWriter.Close(); err != nil {
			return err
		}
		l.pipeWriter = nil
	}
	return nil
}

func DecompressReadCloser(reader io.Reader) *DecompressedReadCloser {
	return &DecompressedReadCloser{reader: reader}
}

type DecompressedReadCloser struct {
	reader  io.Reader
	decoder *zstd.Decoder
}

func (d *DecompressedReadCloser) Read(p []byte) (int, error) {
	if d.decoder == nil {
		if d.reader == nil {
			return 0, errors.New("reader is missing")
		}
		if decoder, err := zstd.NewReader(d.reader); err != nil {
			return 0, errors.Wrap(err, "unable to create decoder")
		} else {
			d.decoder = decoder
		}
	}
	return d.decoder.Read(p)
}

func (d *DecompressedReadCloser) Close() error {
	if d.decoder != nil {
		d.decoder.Close()
	}
	return nil
}

func SizeReader(reader io.Reader) *SizedReader {
	return &SizedReader{reader: reader}
}

type SizedReader struct {
	reader io.Reader
	size   int64
}

func (s *SizedReader) Read(p []byte) (int, error) {
	if s.reader == nil {
		return 0, errors.New("reader is missing")
	}
	n, err := s.reader.Read(p)
	s.size += int64(n)
	return n, err
}

func (s *SizedReader) Size() int64 {
	return s.size
}

func HeadReader(reader io.Reader, limit int) *HeadedReader {
	return &HeadedReader{reader: reader, limit: limit}
}

type HeadedReader struct {
	reader io.Reader
	limit  int
	buffer *bytes.Buffer
}

func (h *HeadedReader) Read(p []byte) (int, error) {
	if h.buffer == nil {
		if h.reader == nil {
			return 0, errors.New("reader is missing")
		}
		h.buffer = &bytes.Buffer{}
	}
	n, err := h.reader.Read(p)
	_, _ = h.buffer.Write(p[:max(0, min(n, h.limit-h.buffer.Len()))]) // Never returns error
	return n, err
}

func (h *HeadedReader) Head() []byte {
	if h.buffer != nil {
		return h.buffer.Bytes()
	} else {
		return nil
	}
}

func JSONEncoderReader(data any) io.Reader {
	if data == nil {
		return bytes.NewReader(nil)
	}

	pipeReader, pipeWriter := io.Pipe()

	go func() {
		if err := json.NewEncoder(pipeWriter).Encode(data); err != nil {
			_ = pipeWriter.CloseWithError(err)
		} else {
			_ = pipeWriter.Close()
		}
	}()

	return pipeReader
}
