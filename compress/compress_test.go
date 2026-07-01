package compress_test

import (
	"bytes"
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/klauspost/compress/zstd"

	"github.com/tidepool-org/platform/compress"
	compressTest "github.com/tidepool-org/platform/compress/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Compress", func() {
	Context("CompressReadCloser", func() {
		It("returns a reader with limit zero", func() {
			Expect(compress.CompressReadCloser(bytes.NewReader(test.RandomBytes()))).ToNot(BeNil())
		})
	})

	Context("CompressedReadCloser", func() {
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
				var reader *compress.CompressedReadCloser

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
				compressed := compressTest.Compress(originalData)
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
				reader := compress.DecompressReadCloser(bytes.NewReader(compressTest.Compress(originalData)))
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
		Context("Read", func() {
			It("returns an error when there is no reader", func() {
				reader := compress.SizeReader(nil)
				buffer := make([]byte, 4)
				n, err := reader.Read(buffer)
				Expect(err).To(MatchError("reader is missing"))
				Expect(n).To(BeZero())
			})

			It("returns size read after reading all data", func() {
				data := test.RandomBytes()
				reader := compress.SizeReader(bytes.NewReader(data))
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Size()).To(Equal(int64(len(data))))
			})

			It("accumulates size correctly across multiple reads", func() {
				data := test.RandomBytes()
				reader := compress.SizeReader(bytes.NewReader(data))
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
				reader := compress.SizeReader(test.ErrorReader(testErr))
				buffer := make([]byte, 4)
				n, err := reader.Read(buffer)
				Expect(n).To(BeZero())
				Expect(err).To(Equal(testErr))
				Expect(reader.Size()).To(BeZero())
			})
		})

		Context("Size", func() {
			It("returns zero before any Read", func() {
				reader := compress.SizeReader(nil)
				Expect(reader.Size()).To(BeZero())
			})

			It("returns size after reading all data", func() {
				data := test.RandomBytes()
				reader := compress.SizeReader(bytes.NewReader(data))
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Size()).To(Equal(int64(len(data))))
			})
		})
	})

	Context("HeadReader", func() {
		It("returns a HeadedReader with the specified limit", func() {
			limit := test.RandomInt()
			reader := compress.HeadReader(bytes.NewReader(nil), limit)
			Expect(reader).ToNot(BeNil())
		})
	})

	Context("HeadedReader", func() {
		Context("Read", func() {
			It("returns an error when there is no reader", func() {
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
				Expect(reader.Head()).To(Equal(data))
			})

			It("captures all bytes when data is exactly equal to the limit", func() {
				limit := test.RandomIntFromRange(10, 100)
				data := test.RandomBytesFromRange(limit, limit)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Head()).To(Equal(data))
			})

			It("captures exactly limit bytes when data is longer than the limit", func() {
				limit := test.RandomIntFromRange(10, 100)
				data := test.RandomBytesFromRange(limit, 1000)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Head()).To(Equal(data[:limit]))
			})

			It("propagates errors from the underlying reader", func() {
				testErr := errorsTest.RandomError()
				reader := compress.HeadReader(test.ErrorReader(testErr), test.RandomIntFromRange(100, 1000))
				buffer := make([]byte, test.RandomIntFromRange(10, 100))
				n, err := reader.Read(buffer)
				Expect(err).To(Equal(testErr))
				Expect(n).To(BeZero())
				Expect(reader.Head()).To(BeEmpty())
			})

			It("captures no bytes when the limit is zero", func() {
				data := test.RandomBytes()
				reader := compress.HeadReader(bytes.NewReader(data), 0)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Head()).To(BeEmpty())
			})
		})

		Context("Head", func() {
			It("returns nil before any Read", func() {
				reader := compress.HeadReader(bytes.NewReader(test.RandomBytes()), test.RandomInt())
				Expect(reader.Head()).To(BeNil())
			})

			It("returns captured head bytes after reading", func() {
				limit := test.RandomIntFromRange(10, 100)
				data := test.RandomBytesFromRange(limit, 1000)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Head()).To(Equal(data[:limit]))
			})

			It("returns all bytes when data is shorter than limit", func() {
				limit := test.RandomIntFromRange(100, 1000)
				data := test.RandomBytesFromRange(10, limit)
				reader := compress.HeadReader(bytes.NewReader(data), limit)
				Expect(io.ReadAll(reader)).To(Equal(data))
				Expect(reader.Head()).To(Equal(data))
			})
		})
	})

	Context("JSONEncoderReader", func() {
		It("returns an empty reader when data is nil", func() {
			reader := compress.JSONEncoderReader(nil)
			Expect(io.ReadAll(reader)).To(BeEmpty())
		})

		It("JSON encodes an array value", func() {
			input := []string{test.RandomString(), test.RandomString()}
			reader := compress.JSONEncoderReader(input)
			data, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(MatchJSON(fmt.Sprintf(`[%q, %q]`, input[0], input[1])))
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

		It("returns an error when data cannot be JSON-encoded", func() {
			reader := compress.JSONEncoderReader(func() {})
			_, err := io.ReadAll(reader)
			Expect(err).To(HaveOccurred())
		})
	})
})
