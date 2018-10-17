package file_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	storeUnstructuredFile "github.com/tidepool-org/platform/store/unstructured/file"
	storeUnstructuredTest "github.com/tidepool-org/platform/store/unstructured/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("File", func() {
	It("has type file", func() {
		Expect(storeUnstructuredFile.Type).To(Equal("file"))
	})

	Context("with config", func() {
		var directory string
		var cfg *storeUnstructuredFile.Config

		BeforeEach(func() {
			directory = test.RandomTemporaryDirectory()
			cfg = storeUnstructuredFile.NewConfig()
			Expect(cfg).ToNot(BeNil())
			cfg.Directory = directory
		})

		AfterEach(func() {
			if directory != "" {
				Expect(os.RemoveAll(directory)).To(Succeed())
			}
		})

		Context("NewStore", func() {
			It("return an error if the config is missing", func() {
				str, err := storeUnstructuredFile.NewStore(nil)
				Expect(err).To(MatchError("config is missing"))
				Expect(str).To(BeNil())
			})

			It("return an error if the config is invalid", func() {
				cfg.Directory = ""
				str, err := storeUnstructuredFile.NewStore(cfg)
				Expect(err).To(MatchError("config is invalid; directory is missing"))
				Expect(str).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(storeUnstructuredFile.NewStore(cfg)).ToNot(BeNil())
			})
		})

		Context("with new store", func() {
			var str *storeUnstructuredFile.Store
			var ctx context.Context
			var key string
			var keyPath string
			var contents []byte

			BeforeEach(func() {
				var err error
				str, err = storeUnstructuredFile.NewStore(cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(str).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
				key = storeUnstructuredTest.RandomKey()
				keyPath = filepath.Join(directory, filepath.FromSlash(key))
				contents = []byte(test.RandomString())
			})

			Context("Exists", func() {
				It("returns an error if the context is missing", func() {
					exists, err := str.Exists(nil, key)
					Expect(err).To(MatchError("context is missing"))
					Expect(exists).To(BeFalse())
				})

				It("returns an error if the key is missing", func() {
					exists, err := str.Exists(ctx, "")
					Expect(err).To(MatchError("key is missing"))
					Expect(exists).To(BeFalse())
				})

				It("returns an error if the key is invalid", func() {
					exists, err := str.Exists(ctx, "#invalid#")
					Expect(err).To(MatchError("key is invalid"))
					Expect(exists).To(BeFalse())
				})

				It("returns an error if the key is a directory", func() {
					Expect(os.MkdirAll(keyPath, 0777)).To(Succeed())
					exists, err := str.Exists(ctx, key)
					Expect(err).To(MatchError(fmt.Sprintf("unexpected directory or irregular file at path %q", keyPath)))
					Expect(exists).To(BeFalse())
				})

				It("returns false if the key does not exist", func() {
					Expect(str.Exists(ctx, key)).To(BeFalse())
				})

				It("returns true if the key exists", func() {
					Expect(os.MkdirAll(filepath.Dir(keyPath), 0777)).To(Succeed())
					Expect(ioutil.WriteFile(keyPath, contents, 0666)).To(Succeed())
					Expect(str.Exists(ctx, key)).To(BeTrue())
				})
			})

			Context("Put", func() {
				var reader io.Reader

				BeforeEach(func() {
					reader = bytes.NewReader(contents)
				})

				It("returns an error if the context is missing", func() {
					Expect(str.Put(nil, key, reader)).To(MatchError("context is missing"))
				})

				It("returns an error if the key is missing", func() {
					Expect(str.Put(ctx, "", reader)).To(MatchError("key is missing"))
				})

				It("returns an error if the key is invalid", func() {
					Expect(str.Put(ctx, "#invalid#", reader)).To(MatchError("key is invalid"))
				})

				It("returns an error if the reader is missing", func() {
					Expect(str.Put(ctx, key, nil)).To(MatchError("reader is missing"))
				})

				It("returns an error if it is unable to create the directories", func() {
					Expect(os.MkdirAll(filepath.Dir(keyPath), 0777)).To(Succeed())
					Expect(ioutil.WriteFile(keyPath, contents, 0666)).To(Succeed())
					key = path.Join(key, storeUnstructuredTest.RandomKeySegment())
					Expect(str.Put(ctx, key, reader)).To(MatchError(fmt.Sprintf("unable to create directories at path %q; mkdir %s: not a directory", keyPath, keyPath)))
				})

				It("returns an error if it is unable to create the file", func() {
					Expect(os.MkdirAll(keyPath, 0777)).To(Succeed())
					Expect(str.Put(ctx, key, reader)).To(MatchError(fmt.Sprintf("unable to create file at path %q; open %s: is a directory", keyPath, keyPath)))
				})

				It("returns an error if it is unable to write the file", func() {
					err := errorsTest.RandomError()
					rdr := test.NewReader()
					rdr.ReadOutput = &test.ReadOutput{BytesRead: 0, Error: err}
					Expect(str.Put(ctx, key, rdr)).To(MatchError(fmt.Sprintf("unable to write file at path %q; %s", keyPath, err)))
				})

				It("creates file with contents and returns successfully", func() {
					Expect(str.Put(ctx, key, reader)).To(Succeed())
					Expect(keyPath).To(BeARegularFile())
					Expect(ioutil.ReadFile(keyPath)).To(Equal(contents))
				})
			})

			Context("Get", func() {
				var reader io.ReadCloser

				BeforeEach(func() {
					reader = nil
				})

				AfterEach(func() {
					if reader != nil {
						Expect(reader.Close()).To(Succeed())
					}
				})

				It("returns an error if the context is missing", func() {
					var err error
					reader, err = str.Get(nil, key)
					Expect(err).To(MatchError("context is missing"))
					Expect(reader).To(BeNil())
				})

				It("returns an error if the key is missing", func() {
					var err error
					reader, err = str.Get(ctx, "")
					Expect(err).To(MatchError("key is missing"))
					Expect(reader).To(BeNil())
				})

				It("returns an error if the key is invalid", func() {
					var err error
					reader, err = str.Get(ctx, "#invalid#")
					Expect(err).To(MatchError("key is invalid"))
					Expect(reader).To(BeNil())
				})

				It("returns an error if the key is a directory", func() {
					Expect(os.MkdirAll(keyPath, 0777)).To(Succeed())
					var err error
					reader, err = str.Get(ctx, key)
					Expect(err).To(MatchError(fmt.Sprintf("unexpected directory or irregular file at path %q", keyPath)))
					Expect(reader).To(BeNil())
				})

				It("returns no reader if the key does not exist", func() {
					var err error
					reader, err = str.Get(ctx, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).To(BeNil())
				})

				It("returns a reader to content if the key exists", func() {
					Expect(os.MkdirAll(filepath.Dir(keyPath), 0777)).To(Succeed())
					Expect(ioutil.WriteFile(keyPath, contents, 0666)).To(Succeed())
					var err error
					reader, err = str.Get(ctx, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).ToNot(BeNil())
					Expect(ioutil.ReadAll(reader)).To(Equal(contents))
				})
			})

			Context("Delete", func() {
				It("returns an error if the context is missing", func() {
					deleted, err := str.Delete(nil, key)
					Expect(err).To(MatchError("context is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error if the key is missing", func() {
					deleted, err := str.Delete(ctx, "")
					Expect(err).To(MatchError("key is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error if the key is invalid", func() {
					deleted, err := str.Delete(ctx, "#invalid#")
					Expect(err).To(MatchError("key is invalid"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error if the key is a directory", func() {
					Expect(os.MkdirAll(keyPath, 0777)).To(Succeed())
					deleted, err := str.Delete(ctx, key)
					Expect(err).To(MatchError(fmt.Sprintf("unexpected directory or irregular file at path %q", keyPath)))
					Expect(deleted).To(BeFalse())
				})

				It("returns false if the key does not exist", func() {
					Expect(str.Delete(ctx, key)).To(BeFalse())
				})

				Context("with existing file", func() {
					JustBeforeEach(func() {
						Expect(os.MkdirAll(filepath.Dir(keyPath), 0777)).To(Succeed())
						Expect(ioutil.WriteFile(keyPath, contents, 0666)).To(Succeed())
					})

					It("returns true if the key exists and it deletes the file", func() {
						Expect(str.Delete(ctx, key)).To(BeTrue())
						Expect(keyPath).ToNot(BeAnExistingFile())
					})

					Context("with at least one directory", func() {
						BeforeEach(func() {
							key = path.Join(key, storeUnstructuredTest.RandomKeySegment())
							keyPath = filepath.Join(directory, filepath.FromSlash(key))
						})

						It("returns true if the key exists and it deletes the file and any empty directories", func() {
							Expect(str.Delete(ctx, key)).To(BeTrue())
							Expect(keyPath).ToNot(BeAnExistingFile())
							Expect(filepath.Dir(keyPath)).ToNot(BeADirectory())
						})

						It("returns true if the key exists and it deletes the file, but not any non-empty directories", func() {
							Expect(ioutil.WriteFile(keyPath+".exists", contents, 0666)).To(Succeed())
							Expect(str.Delete(ctx, key)).To(BeTrue())
							Expect(keyPath).ToNot(BeAnExistingFile())
							Expect(filepath.Dir(keyPath)).To(BeADirectory())
						})
					})
				})
			})
		})
	})
})
