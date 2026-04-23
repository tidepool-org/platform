package test

import (
	"github.com/klauspost/compress/zstd"

	"github.com/tidepool-org/platform/compress"
)

func Compress(input []byte) []byte {
	return encoder.EncodeAll(input, make([]byte, 0, len(input)))
}

func Decompress(input []byte) ([]byte, error) {
	return decoder.DecodeAll(input, nil)
}

var (
	encoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(compress.CompressionLevel))
	decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
)
