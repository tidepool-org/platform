package compress

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/klauspost/compress/zstd"

	"github.com/tidepool-org/platform/errors"
)

const ErrorCodeLimitExceeded = "limit-exceeded"

func ErrorLimitExceeded(limit int64) error {
	return errors.Preparedf(ErrorCodeLimitExceeded, "limit exceeded", "limit %d exceeded", limit)
}

func CompressReadCloser(reader io.Reader) *LimitedCompressReadCloser {
	return LimitCompressReadCloser(reader, 0)
}

// limit of <= 0 means no limit
func LimitCompressReadCloser(reader io.Reader, limit int64) *LimitedCompressReadCloser {
	return &LimitedCompressReadCloser{reader: reader, limit: limit}
}

type LimitedCompressReadCloser struct {
	reader        io.Reader
	limit         int64
	pipeReader    *io.PipeReader
	pipeWriter    *io.PipeWriter
	limitedWriter *LimitedWriter
}

func (l *LimitedCompressReadCloser) Read(p []byte) (int, error) {
	if l.pipeReader == nil || l.pipeWriter == nil || l.limitedWriter == nil {
		if l.reader == nil {
			return 0, errors.New("reader is missing")
		}

		l.pipeReader, l.pipeWriter = io.Pipe()
		l.limitedWriter = &LimitedWriter{writer: l.pipeWriter, limit: l.limit}

		go func() {
			if encoder, err := zstd.NewWriter(l.limitedWriter, zstd.WithEncoderLevel(zstd.SpeedDefault)); err != nil {
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

func (l *LimitedCompressReadCloser) Close() error {
	if l.pipeReader != nil {
		if err := l.pipeReader.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (l *LimitedCompressReadCloser) Limit() int64 {
	return l.limit
}

func (l *LimitedCompressReadCloser) Size() int64 {
	if l.limitedWriter != nil {
		return l.limitedWriter.Size()
	} else {
		return 0
	}
}

func DecompressReadCloser(reader io.Reader) *DecompressedReadCloser {
	return &DecompressedReadCloser{reader: reader}
}

type DecompressedReadCloser struct {
	reader  io.Reader
	decoder *zstd.Decoder
}

func (d *DecompressedReadCloser) Read(p []byte) (int, error) {
	if err := d.decode(); err != nil {
		return 0, err
	}
	return d.decoder.Read(p)
}

func (d *DecompressedReadCloser) Close() error {
	if d.decoder != nil {
		d.decoder.Close()
	}
	return nil
}

func (d *DecompressedReadCloser) decoded() bool {
	return d.decoder != nil
}

func (d *DecompressedReadCloser) decode() error {
	if !d.decoded() {
		if d.reader == nil {
			return errors.New("reader is missing")
		}
		if decoder, err := zstd.NewReader(d.reader); err != nil {
			return errors.Wrap(err, "unable to create decoder")
		} else {
			d.decoder = decoder
		}
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
	if _, writeErr := h.buffer.Write(p[:min(n, h.limit-h.buffer.Len())]); writeErr != nil && err == nil {
		err = writeErr
	}
	return n, err
}

func (h *HeadedReader) Limit() int {
	return h.limit
}

func (h *HeadedReader) Size() int {
	if h.buffer != nil {
		return h.buffer.Len()
	} else {
		return 0
	}
}

func (h *HeadedReader) Bytes() []byte {
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
		if bites, err := json.Marshal(data); err != nil {
			_ = pipeWriter.CloseWithError(err)
		} else if _, err := pipeWriter.Write(bites); err != nil {
			_ = pipeWriter.CloseWithError(err)
		} else {
			_ = pipeWriter.Close()
		}
	}()

	return pipeReader
}

// limit of <= 0 means no limit
func LimitWriter(writer io.Writer, limit int64) *LimitedWriter {
	return &LimitedWriter{writer: writer, limit: limit}
}

// limit of <= 0 means no limit
type LimitedWriter struct {
	writer io.Writer
	limit  int64
	size   int64
}

func (l *LimitedWriter) Write(p []byte) (int, error) {
	if l.writer == nil {
		return 0, errors.New("writer is missing")
	}
	if l.limit > 0 && l.size+int64(len(p)) > l.limit {
		return 0, ErrorLimitExceeded(l.limit)
	}
	n, err := l.writer.Write(p)
	l.size += int64(n)
	return n, err
}

func (l *LimitedWriter) Limit() int64 {
	return l.limit
}

func (l *LimitedWriter) Size() int64 {
	return l.size
}
