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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
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
		var config *storeUnstructuredFile.Config

		BeforeEach(func() {
			directory = test.RandomTemporaryDirectory()
			config = storeUnstructuredFile.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Directory = directory
		})

		AfterEach(func() {
			if directory != "" {
				Expect(os.RemoveAll(directory)).To(Succeed())
			}
		})

		Context("NewStore", func() {
			It("return an error if the config is missing", func() {
				store, err := storeUnstructuredFile.NewStore(nil)
				Expect(err).To(MatchError("config is missing"))
				Expect(store).To(BeNil())
			})

			It("return an error if the config is invalid", func() {
				config.Directory = ""
				store, err := storeUnstructuredFile.NewStore(config)
				Expect(err).To(MatchError("config is invalid; directory is missing"))
				Expect(store).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(storeUnstructuredFile.NewStore(config)).ToNot(BeNil())
			})
		})

		Context("with new store", func() {
			var store *storeUnstructuredFile.Store
			var ctx context.Context
			var key string
			var keyPath string
			var contents []byte

			BeforeEach(func() {
				var err error
				store, err = storeUnstructuredFile.NewStore(config)
				Expect(err).ToNot(HaveOccurred())
				Expect(store).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				key = storeUnstructuredTest.RandomKey()
				keyPath = filepath.Join(directory, filepath.FromSlash(key))
				contents = []byte(test.RandomString())
			})

			Context("Exists", func() {
				It("returns an error if the context is missing", func() {
					exists, err := store.Exists(nil, key)
					Expect(err).To(MatchError("context is missing"))
					Expect(exists).To(BeFalse())
				})

				It("returns an error if the key is missing", func() {
					exists, err := store.Exists(ctx, "")
					Expect(err).To(MatchError("key is missing"))
					Expect(exists).To(BeFalse())
				})

				It("returns an error if the key is invalid", func() {
					exists, err := store.Exists(ctx, "#invalid#")
					Expect(err).To(MatchError("key is invalid"))
					Expect(exists).To(BeFalse())
				})

				It("returns an error if the key is a directory", func() {
					Expect(os.MkdirAll(keyPath, 0777)).To(Succeed())
					exists, err := store.Exists(ctx, key)
					Expect(err).To(MatchError(fmt.Sprintf("unexpected directory or irregular file at path %q", keyPath)))
					Expect(exists).To(BeFalse())
				})

				It("returns false if the key does not exist", func() {
					Expect(store.Exists(ctx, key)).To(BeFalse())
				})

				It("returns true if the key exists", func() {
					Expect(os.MkdirAll(filepath.Dir(keyPath), 0777)).To(Succeed())
					Expect(ioutil.WriteFile(keyPath, contents, 0666)).To(Succeed())
					Expect(store.Exists(ctx, key)).To(BeTrue())
				})
			})

			Context("Put", func() {
				var reader io.Reader
				var options *storeUnstructured.Options

				BeforeEach(func() {
					reader = bytes.NewReader(contents)
					options = storeUnstructuredTest.RandomOptions()
				})

				It("returns an error if the context is missing", func() {
					Expect(store.Put(nil, key, reader, options)).To(MatchError("context is missing"))
				})

				It("returns an error if the key is missing", func() {
					Expect(store.Put(ctx, "", reader, options)).To(MatchError("key is missing"))
				})

				It("returns an error if the key is invalid", func() {
					Expect(store.Put(ctx, "#invalid#", reader, options)).To(MatchError("key is invalid"))
				})

				It("returns an error if the reader is missing", func() {
					Expect(store.Put(ctx, key, nil, options)).To(MatchError("reader is missing"))
				})

				It("returns an error if the options is invalid", func() {
					options.MediaType = pointer.FromString("")
					Expect(store.Put(ctx, key, reader, options)).To(MatchError("options is invalid; value is empty"))
				})

				It("returns an error if it is unable to create the directories", func() {
					Expect(os.MkdirAll(filepath.Dir(keyPath), 0777)).To(Succeed())
					Expect(ioutil.WriteFile(keyPath, contents, 0666)).To(Succeed())
					key = path.Join(key, storeUnstructuredTest.RandomKeySegment())
					Expect(store.Put(ctx, key, reader, options)).To(MatchError(fmt.Sprintf("unable to create directories at path %q; mkdir %s: not a directory", keyPath, keyPath)))
				})

				It("returns an error if it is unable to create the file", func() {
					Expect(os.MkdirAll(keyPath, 0777)).To(Succeed())
					Expect(store.Put(ctx, key, reader, options)).To(MatchError(fmt.Sprintf("unable to create file at path %q; open %s: is a directory", keyPath, keyPath)))
				})

				It("returns an error if it is unable to write the file", func() {
					err := errorsTest.RandomError()
					rdr := test.NewReader()
					rdr.ReadOutput = &test.ReadOutput{BytesRead: 0, Error: err}
					Expect(store.Put(ctx, key, rdr, options)).To(MatchError(fmt.Sprintf("unable to write file at path %q; %s", keyPath, err)))
				})

				It("creates file with contents and returns successfully", func() {
					Expect(store.Put(ctx, key, reader, options)).To(Succeed())
					Expect(keyPath).To(BeARegularFile())
					Expect(ioutil.ReadFile(keyPath)).To(Equal(contents))
				})

				It("creates file with contents and returns successfully without options", func() {
					options = nil
					Expect(store.Put(ctx, key, reader, options)).To(Succeed())
					Expect(keyPath).To(BeARegularFile())
					Expect(ioutil.ReadFile(keyPath)).To(Equal(contents))
				})

				It("creates file with contents and returns successfully with options with media type missing", func() {
					options.MediaType = nil
					Expect(store.Put(ctx, key, reader, options)).To(Succeed())
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
					reader, err = store.Get(nil, key)
					Expect(err).To(MatchError("context is missing"))
					Expect(reader).To(BeNil())
				})

				It("returns an error if the key is missing", func() {
					var err error
					reader, err = store.Get(ctx, "")
					Expect(err).To(MatchError("key is missing"))
					Expect(reader).To(BeNil())
				})

				It("returns an error if the key is invalid", func() {
					var err error
					reader, err = store.Get(ctx, "#invalid#")
					Expect(err).To(MatchError("key is invalid"))
					Expect(reader).To(BeNil())
				})

				It("returns an error if the key is a directory", func() {
					Expect(os.MkdirAll(keyPath, 0777)).To(Succeed())
					var err error
					reader, err = store.Get(ctx, key)
					Expect(err).To(MatchError(fmt.Sprintf("unexpected directory or irregular file at path %q", keyPath)))
					Expect(reader).To(BeNil())
				})

				It("returns no reader if the key does not exist", func() {
					var err error
					reader, err = store.Get(ctx, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).To(BeNil())
				})

				It("returns a reader to content if the key exists", func() {
					Expect(os.MkdirAll(filepath.Dir(keyPath), 0777)).To(Succeed())
					Expect(ioutil.WriteFile(keyPath, contents, 0666)).To(Succeed())
					var err error
					reader, err = store.Get(ctx, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).ToNot(BeNil())
					Expect(ioutil.ReadAll(reader)).To(Equal(contents))
				})
			})

			Context("Delete", func() {
				It("returns an error if the context is missing", func() {
					deleted, err := store.Delete(nil, key)
					Expect(err).To(MatchError("context is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error if the key is missing", func() {
					deleted, err := store.Delete(ctx, "")
					Expect(err).To(MatchError("key is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error if the key is invalid", func() {
					deleted, err := store.Delete(ctx, "#invalid#")
					Expect(err).To(MatchError("key is invalid"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error if the key is a directory", func() {
					Expect(os.MkdirAll(keyPath, 0777)).To(Succeed())
					deleted, err := store.Delete(ctx, key)
					Expect(err).To(MatchError(fmt.Sprintf("unexpected directory or irregular file at path %q", keyPath)))
					Expect(deleted).To(BeFalse())
				})

				It("returns false if the key does not exist", func() {
					Expect(store.Delete(ctx, key)).To(BeFalse())
				})

				Context("with existing file", func() {
					JustBeforeEach(func() {
						Expect(os.MkdirAll(filepath.Dir(keyPath), 0777)).To(Succeed())
						Expect(ioutil.WriteFile(keyPath, contents, 0666)).To(Succeed())
					})

					It("returns true if the key exists and it deletes the file", func() {
						Expect(store.Delete(ctx, key)).To(BeTrue())
						Expect(keyPath).ToNot(BeAnExistingFile())
					})

					Context("with at least one directory", func() {
						BeforeEach(func() {
							key = path.Join(key, storeUnstructuredTest.RandomKeySegment())
							keyPath = filepath.Join(directory, filepath.FromSlash(key))
						})

						It("returns true if the key exists and it deletes the file and any empty directories", func() {
							Expect(store.Delete(ctx, key)).To(BeTrue())
							Expect(keyPath).ToNot(BeAnExistingFile())
							Expect(filepath.Dir(keyPath)).ToNot(BeADirectory())
						})

						It("returns true if the key exists and it deletes the file, but not any non-empty directories", func() {
							Expect(ioutil.WriteFile(keyPath+".exists", contents, 0666)).To(Succeed())
							Expect(store.Delete(ctx, key)).To(BeTrue())
							Expect(keyPath).ToNot(BeAnExistingFile())
							Expect(filepath.Dir(keyPath)).To(BeADirectory())
						})
					})
				})
			})

			Context("DeleteDirectory", func() {
				It("returns an error if the context is missing", func() {
					Expect(store.DeleteDirectory(nil, key)).To(MatchError("context is missing"))
				})

				It("returns an error if the key is missing", func() {
					Expect(store.DeleteDirectory(ctx, "")).To(MatchError("key is missing"))
				})

				It("returns an error if the key is invalid", func() {
					Expect(store.DeleteDirectory(ctx, "#invalid#")).To(MatchError("key is invalid"))
				})

				It("returns an error if the key is not a directory", func() {
					Expect(os.MkdirAll(filepath.Dir(keyPath), 0777)).To(Succeed())
					Expect(ioutil.WriteFile(keyPath, contents, 0666)).To(Succeed())
					Expect(store.DeleteDirectory(ctx, key)).To(MatchError(fmt.Sprintf("unexpected file at path %q", keyPath)))
				})

				It("returns successfully if the key does not exist", func() {
					Expect(store.DeleteDirectory(ctx, key)).To(Succeed())
				})

				Context("with existing directory", func() {
					JustBeforeEach(func() {
						Expect(os.MkdirAll(keyPath, 0777)).To(Succeed())
						Expect(ioutil.WriteFile(filepath.Join(keyPath, "exists"), contents, 0666)).To(Succeed())
					})

					It("returns successfully if the key exists and it deletes the file", func() {
						Expect(store.DeleteDirectory(ctx, key)).To(Succeed())
						Expect(keyPath).ToNot(BeAnExistingFile())
					})

					Context("with at least one directory", func() {
						BeforeEach(func() {
							key = path.Join(key, storeUnstructuredTest.RandomKeySegment())
							keyPath = filepath.Join(directory, filepath.FromSlash(key))
						})

						It("returns successfully if the key exists and it deletes the file and any empty directories", func() {
							Expect(store.DeleteDirectory(ctx, key)).To(Succeed())
							Expect(keyPath).ToNot(BeAnExistingFile())
							Expect(filepath.Dir(keyPath)).ToNot(BeADirectory())
						})

						It("returns successfully if the key exists and it deletes the file, but not any non-empty directories", func() {
							Expect(ioutil.WriteFile(keyPath+".exists", contents, 0666)).To(Succeed())
							Expect(store.DeleteDirectory(ctx, key)).To(Succeed())
							Expect(keyPath).ToNot(BeAnExistingFile())
							Expect(filepath.Dir(keyPath)).To(BeADirectory())
						})
					})
				})
			})
		})
	})
})
