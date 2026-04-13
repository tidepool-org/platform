package compress_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/klauspost/compress/zstd"

	"github.com/tidepool-org/platform/compress"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Compress", func() {
	It("ErrorCodeLimitExceeded is expected", func() {
		Expect(compress.ErrorCodeLimitExceeded).To(Equal("limit-exceeded"))
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorLimitExceeded", compress.ErrorLimitExceeded(1<<30), "limit-exceeded", "limit exceeded", fmt.Sprintf("limit %d exceeded", int64(1<<30))),
		)
	})

	Context("CompressReadCloser", func() {
		It("returns a reader with limit zero", func() {
			reader := compress.CompressReadCloser(bytes.NewReader(test.RandomBytes()))
			Expect(reader).ToNot(BeNil())
			Expect(reader.Limit()).To(BeZero())
		})
	})

	Context("LimitCompressReadCloser", func() {
		It("returns a reader with the specified limit", func() {
			limit := test.RandomInt64()
			reader := compress.LimitCompressReadCloser(bytes.NewReader(test.RandomBytes()), limit)
			Expect(reader).ToNot(BeNil())
			Expect(reader.Limit()).To(Equal(limit))
		})

		It("returns a reader with limit zero when given zero", func() {
			reader := compress.LimitCompressReadCloser(bytes.NewReader(nil), 0)
			Expect(reader).ToNot(BeNil())
			Expect(reader.Limit()).To(BeZero())
		})
	})

	Context("LimitedCompressReadCloser", func() {
		Context("Read", func() {
			It("returns an error when reader is nil", func() {
				reader := compress.CompressReadCloser(nil)
				buffer := make([]byte, 64)
				n, err := reader.Read(buffer)
				Expect(n).To(BeZero())
				Expect(err).To(MatchError("reader is missing"))
			})

			Context("with valid reader", func() {
				var originalData []byte
				var reader *compress.LimitedCompressReadCloser

				BeforeEach(func() {
					originalData = test.RandomBytes()
					reader = compress.CompressReadCloser(bytes.NewReader(originalData))
				})

				AfterEach(func() {
					if reader != nil {
						reader.Close()
					}
				})

				It("compresses data successfully", func() {
					compressedData, err := io.ReadAll(reader)
					Expect(err).ToNot(HaveOccurred())
					Expect(compressedData).ToNot(BeEmpty())
					Expect(reader.Size()).To(Equal(int64(len(compressedData))))
				})

				It("compressed data round-trips back to original", func() {
					compressedData, err := io.ReadAll(reader)
					Expect(err).ToNot(HaveOccurred())
					Expect(compressedData).ToNot(BeEmpty())
					decoder, err := zstd.NewReader(bytes.NewReader(compressedData))
					Expect(err).ToNot(HaveOccurred())
					Expect(decoder).ToNot(BeNil())
					defer decoder.Close()
					Expect(io.ReadAll(decoder)).To(Equal(originalData))
				})

				It("can be read in multiple chunks", func() {
					buffer := make([]byte, 4)
					var compressedData []byte
					for {
						n, err := reader.Read(buffer)
						if errors.Is(err, io.EOF) {
							break
						}
						Expect(err).ToNot(HaveOccurred())
						compressedData = append(compressedData, buffer[:n]...)
					}
					decoder, err := zstd.NewReader(bytes.NewReader(compressedData))
					Expect(err).ToNot(HaveOccurred())
					Expect(decoder).ToNot(BeNil())
					defer decoder.Close()
					Expect(io.ReadAll(decoder)).To(Equal(originalData))
				})

				It("returns a limit exceeded error with limit of 1", func() {
					limit := test.RandomInt64FromRange(1, 10)
					reader = compress.LimitCompressReadCloser(bytes.NewReader(originalData), limit)
					_, err := io.ReadAll(reader)
					Expect(err).To(HaveOccurred())
					Expect(errors.Code(err)).To(Equal(compress.ErrorCodeLimitExceeded))
				})

				It("succeeds with limit equal to compressed size", func() {
					compressedData, err := io.ReadAll(reader)
					Expect(err).ToNot(HaveOccurred())
					Expect(compressedData).ToNot(BeEmpty())
					compressedLength := int64(len(compressedData))

					reader = compress.LimitCompressReadCloser(bytes.NewReader(originalData), compressedLength)
					Expect(io.ReadAll(reader)).To(Equal(compressedData))
				})

				Context("Close", func() {
					It("succeeds before any read", func() {
						Expect(reader.Close()).To(Succeed())
					})

					It("succeeds after reading all data", func() {
						_, err := io.ReadAll(reader)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader.Close()).To(Succeed())
					})
				})

				Context("Limit", func() {
					It("returns the specified limit", func() {
						limit := test.RandomInt64()
						reader = compress.LimitCompressReadCloser(bytes.NewReader(originalData), limit)
						Expect(reader.Limit()).To(Equal(limit))
					})
				})

				Context("Size", func() {
					It("returns zero before any read", func() {
						Expect(reader.Size()).To(BeZero())
					})

					It("returns the compressed byte count after reading all data", func() {
						compressedData, err := io.ReadAll(reader)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader.Size()).To(Equal(int64(len(compressedData))))
					})
				})
			})
		})
	})

	Context("DecompressReadCloser", func() {
		It("returns a DecompressedReadCloser", func() {
			Expect(compress.DecompressReadCloser(bytes.NewReader(nil))).ToNot(BeNil())
		})
	})

	Context("DecompressedReadCloser", func() {
		Context("Read", func() {
			Context("when reader is nil", func() {
				It("returns an error", func() {
					reader := compress.DecompressReadCloser(nil)
					buffer := make([]byte, 64)
					n, err := reader.Read(buffer)
					Expect(n).To(BeZero())
					Expect(err).To(MatchError("reader is missing"))
				})
			})

			It("decompresses valid compressed data to the original", func() {
				originalData := test.RandomBytes()
				compressed := compressBytes(originalData)
				reader := compress.DecompressReadCloser(bytes.NewReader(compressed))
				Expect(io.ReadAll(reader)).To(Equal(originalData))
			})

			It("returns an error for invalid compressed data", func() {
				reader := compress.DecompressReadCloser(bytes.NewReader([]byte("not zstd data")))
				buffer := make([]byte, 64)
				_, err := reader.Read(buffer)
				Expect(err).To(HaveOccurred())
			})

			It("round-trips correctly when chained with CompressReadCloser", func() {
				originalData := test.RandomBytes()
				reader := compress.DecompressReadCloser(compress.CompressReadCloser(bytes.NewReader(originalData)))
				Expect(io.ReadAll(reader)).To(Equal(originalData))
			})
		})

		Context("Close", func() {
			It("succeeds before any Read", func() {
				reader := compress.DecompressReadCloser(bytes.NewReader(nil))
				Expect(reader.Close()).To(Succeed())
			})

			It("succeeds after reading all data", func() {
				originalData := test.RandomBytes()
				reader := compress.DecompressReadCloser(bytes.NewReader(compressBytes(originalData)))
				Expect(io.ReadAll(reader)).To(Equal(originalData))
				Expect(reader.Close()).To(Succeed())
			})
		})
	})

	Context("SizeReader", func() {
		It("returns a SizedReader", func() {
			reader := compress.SizeReader(bytes.NewReader(nil))
			Expect(reader).ToNot(BeNil())
		})
	})

	Context("SizedReader", func() {
		var data []byte
		var reader *compress.SizedReader

		BeforeEach(func() {
			data = test.RandomBytes()
			reader = compress.SizeReader(bytes.NewReader(data))
		})

		Context("Size", func() {
			It("returns zero before any Read", func() {
				Expect(reader.Size()).To(BeZero())
			})

			It("returns size after reading all data", func() {
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Size()).To(Equal(int64(len(data))))
			})
		})

		Context("Read", func() {
			It("returns size read after reading all data", func() {
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Size()).To(Equal(int64(len(data))))
			})

			It("accumulates size correctly across multiple reads", func() {
				buffer := make([]byte, 4)
				total := 0
				for {
					n, err := reader.Read(buffer)
					total += n
					if errors.Is(err, io.EOF) {
						break
					}
					Expect(err).ToNot(HaveOccurred())
					Expect(reader.Size()).To(Equal(int64(total)))
				}
				Expect(total).To(Equal(len(data)))
				Expect(reader.Size()).To(Equal(int64(total)))
			})

			It("propagates errors from the underlying reader", func() {
				testErr := errorsTest.RandomError()
				reader = compress.SizeReader(test.ErrorReader(testErr))
				buffer := make([]byte, 4)
				n, err := reader.Read(buffer)
				Expect(n).To(BeZero())
				Expect(err).To(Equal(testErr))
				Expect(reader.Size()).To(BeZero())
			})
		})
	})

	Context("HeadReader", func() {
		It("returns a HeadedReader with the specified limit", func() {
			limit := test.RandomInt()
			reader := compress.HeadReader(bytes.NewReader(nil), limit)
			Expect(reader).ToNot(BeNil())
			Expect(reader.Limit()).To(Equal(limit))
		})
	})

	Context("HeadedReader", func() {
		Context("Read", func() {
			It("returns an error", func() {
				reader := compress.HeadReader(nil, test.RandomInt())
				buffer := make([]byte, 4)
				n, err := reader.Read(buffer)
				Expect(err).To(MatchError("reader is missing"))
				Expect(n).To(BeZero())
			})

			It("captures all bytes when data is shorter than the limit", func() {
				limit := test.RandomIntFromRange(100, 1000)
				data := test.RandomBytesFromRange(10, limit)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Bytes()).To(Equal(data))
				Expect(reader.Size()).To(Equal(len(data)))
			})

			It("captures all bytes when data is exactly equal to the limit", func() {
				limit := test.RandomIntFromRange(10, 100)
				data := test.RandomBytesFromRange(limit, limit)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Bytes()).To(Equal(data))
				Expect(reader.Size()).To(Equal(len(data)))
			})

			It("captures exactly limit bytes when data is longer than the limit", func() {
				limit := test.RandomIntFromRange(10, 100)
				data := test.RandomBytesFromRange(limit, 1000)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Bytes()).To(Equal(data[:limit]))
				Expect(reader.Size()).To(Equal(limit))
			})

			It("propagates errors from the underlying reader", func() {
				testErr := errorsTest.RandomError()
				reader := compress.HeadReader(test.ErrorReader(testErr), test.RandomIntFromRange(100, 1000))
				buffer := make([]byte, test.RandomIntFromRange(10, 100))
				n, err := reader.Read(buffer)
				Expect(err).To(Equal(testErr))
				Expect(n).To(BeZero())
				Expect(reader.Bytes()).To(BeEmpty())
				Expect(reader.Size()).To(BeZero())
			})

			It("captures no bytes when the limit is zero", func() {
				data := test.RandomBytes()
				reader := compress.HeadReader(bytes.NewReader(data), 0)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Bytes()).To(BeEmpty())
				Expect(reader.Size()).To(BeZero())
			})
		})

		Context("Limit", func() {
			It("returns the value set at construction", func() {
				limit := test.RandomInt()
				reader := compress.HeadReader(bytes.NewReader(nil), limit)
				Expect(reader.Limit()).To(Equal(limit))
			})
		})

		Context("Size", func() {
			It("returns zero before any Read", func() {
				reader := compress.HeadReader(bytes.NewReader(test.RandomBytes()), test.RandomInt())
				Expect(reader.Size()).To(BeZero())
			})

			It("returns the number of captured head bytes after reading", func() {
				limit := test.RandomIntFromRange(10, 100)
				data := test.RandomBytesFromRange(limit, 1000)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Size()).To(Equal(limit))
			})
		})

		Context("Bytes", func() {
			It("returns nil before any Read", func() {
				reader := compress.HeadReader(bytes.NewReader(test.RandomBytes()), test.RandomInt())
				Expect(reader.Bytes()).To(BeNil())
			})

			It("returns captured head bytes after reading", func() {
				limit := test.RandomIntFromRange(10, 100)
				data := test.RandomBytesFromRange(limit, 1000)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Bytes()).To(Equal(data[:limit]))
			})

			It("returns all bytes when data is shorter than limit", func() {
				limit := test.RandomIntFromRange(100, 1000)
				data := test.RandomBytesFromRange(10, limit)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Bytes()).To(Equal(data))
			})
		})
	})

	Context("JSONEncoderReader", func() {
		It("returns an empty reader when data is nil", func() {
			reader := compress.JSONEncoderReader(nil)
			Expect(io.ReadAll(reader)).To(BeEmpty())
		})

		It("JSON encodes a string value with trailing newline", func() {
			input := test.RandomString()
			reader := compress.JSONEncoderReader(input)
			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal(test.Must(json.Marshal(input))))
		})

		It("JSON encodes a struct value", func() {
			type TestStruct struct {
				Name string `json:"name"`
			}
			input := TestStruct{Name: test.RandomString()}
			reader := compress.JSONEncoderReader(input)
			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(MatchJSON(fmt.Sprintf(`{"name":%q}`, input.Name)))
		})

		It("JSON encodes an integer with trailing newline", func() {
			input := test.RandomInt64()
			reader := compress.JSONEncoderReader(input)
			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(strconv.FormatInt(input, 10)))
		})

		It("returns an error when data cannot be JSON-encoded", func() {
			reader := compress.JSONEncoderReader(func() {})
			_, err := io.ReadAll(reader)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("LimitWriter", func() {
		It("returns a LimitedWriter with the specified limit", func() {
			limit := test.RandomInt64()
			writer := compress.LimitWriter(&bytes.Buffer{}, limit)
			Expect(writer).ToNot(BeNil())
			Expect(writer.Limit()).To(Equal(limit))
		})
	})

	Context("LimitedWriter", func() {
		var baseWriter *bytes.Buffer

		BeforeEach(func() {
			baseWriter = &bytes.Buffer{}
		})

		Context("Write", func() {
			var writer *compress.LimitedWriter

			It("returns an error when the base write is missing", func() {
				writer = compress.LimitWriter(nil, test.RandomInt64())
				n, err := writer.Write(test.RandomBytes())
				Expect(n).To(BeZero())
				Expect(err).To(MatchError("writer is missing"))
			})

			Context("with no limit", func() {
				BeforeEach(func() {
					writer = compress.LimitWriter(baseWriter, 0)
				})

				It("writes any amount of data without error", func() {
					data := test.RandomBytes()
					Expect(writer.Write(data)).To(Equal(len(data)))
					Expect(baseWriter.Bytes()).To(Equal(data))
				})

				It("accumulates size across multiple writes", func() {
					data1 := test.RandomBytes()
					data2 := test.RandomBytes()
					Expect(writer.Write(data1)).To(Equal(len(data1)))
					Expect(writer.Write(data2)).To(Equal(len(data2)))
					Expect(writer.Size()).To(Equal(int64(len(data1) + len(data2))))
				})
			})

			Context("with a limit", func() {
				var limit int

				BeforeEach(func() {
					limit = test.RandomIntFromRange(10, 100)
					writer = compress.LimitWriter(baseWriter, int64(limit))
				})

				It("returns a limit exceeded error before writing when a single write exceeds the limit", func() {
					data := test.RandomBytesFromRange(limit+1, limit+1)
					n, err := writer.Write(data)
					Expect(errors.Code(err)).To(Equal(compress.ErrorCodeLimitExceeded))
					Expect(n).To(BeZero())
					Expect(baseWriter.Len()).To(BeZero())
				})

				It("succeeds when writing exactly limit bytes", func() {
					data := test.RandomBytesFromRange(limit, limit)
					Expect(writer.Write(data)).To(Equal(limit))
					Expect(writer.Size()).To(Equal(int64(limit)))
					Expect(baseWriter.Bytes()).To(Equal(data))
				})

				It("succeeds when writing fewer than limit bytes", func() {
					data := test.RandomBytesFromRange(limit-1, limit-1)
					Expect(writer.Write(data)).To(Equal(limit - 1))
					Expect(writer.Size()).To(Equal(int64(limit - 1)))
					Expect(baseWriter.Bytes()).To(Equal(data))
				})

				It("returns a limit exceeded error when cumulative writes exceed the limit", func() {
					Expect(writer.Write(test.RandomBytesFromRange(limit, limit))).To(Equal(limit))
					n, err := writer.Write(test.RandomBytesFromRange(1, 1))
					Expect(errors.Code(err)).To(Equal(compress.ErrorCodeLimitExceeded))
					Expect(n).To(BeZero())
					Expect(writer.Size()).To(Equal(int64(limit)))
				})

				It("accumulates size correctly across successful writes", func() {
					half := limit / 2
					Expect(writer.Write(test.RandomBytesFromRange(half, half))).To(Equal(half))
					Expect(writer.Size()).To(Equal(int64(half)))
					Expect(writer.Write(test.RandomBytesFromRange(half, half))).To(Equal(half))
					Expect(writer.Size()).To(Equal(int64(half * 2)))
				})
			})
		})

		Context("Limit", func() {
			It("returns zero for an unlimited writer", func() {
				writer := compress.LimitWriter(baseWriter, 0)
				Expect(writer.Limit()).To(BeZero())
			})

			It("returns the specified positive limit", func() {
				limit := test.RandomInt64()
				writer := compress.LimitWriter(baseWriter, limit)
				Expect(writer.Limit()).To(Equal(limit))
			})
		})

		Context("Size", func() {
			It("returns zero before any Write", func() {
				writer := compress.LimitWriter(baseWriter, test.RandomInt64())
				Expect(writer.Size()).To(BeZero())
			})

			It("returns the cumulative bytes written after successful writes", func() {
				data := test.RandomBytes()
				writer := compress.LimitWriter(baseWriter, 0)
				Expect(writer.Write(data)).To(Equal(len(data)))
				Expect(writer.Size()).To(Equal(int64(len(data))))
			})
		})
	})
})

func compressBytes(data []byte) []byte {
	var buffer bytes.Buffer
	encoder, err := zstd.NewWriter(&buffer)
	Expect(err).ToNot(HaveOccurred())
	Expect(encoder).ToNot(BeNil())
	_, err = encoder.Write(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(encoder.Close()).To(Succeed())
	return buffer.Bytes()
}
