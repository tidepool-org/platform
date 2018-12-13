package s3_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	awsSdkAws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	awsTest "github.com/tidepool-org/platform/aws/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	storeUnstructuredS3 "github.com/tidepool-org/platform/store/unstructured/s3"
	storeUnstructuredTest "github.com/tidepool-org/platform/store/unstructured/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("S3", func() {
	It("has type s3", func() {
		Expect(storeUnstructuredS3.Type).To(Equal("s3"))
	})

	Context("with config", func() {
		var config *storeUnstructuredS3.Config
		var awsAPI *awsTest.API
		var awsS3 *awsTest.S3
		var awsS3Manager *awsTest.S3Manager
		var awsUploader *awsTest.Uploader
		var awsDownloader *awsTest.Downloader

		BeforeEach(func() {
			config = storeUnstructuredS3.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Bucket = awsTest.RandomBucket()
			config.Prefix = storeUnstructuredTest.RandomKey()
			awsAPI = awsTest.NewAPI()
			awsS3 = awsTest.NewS3()
			awsAPI.SetS3Output(awsS3)
			awsS3Manager = awsTest.NewS3Manager()
			awsAPI.SetS3ManagerOutput(awsS3Manager)
			awsUploader = awsTest.NewUploader()
			awsS3Manager.SetUploaderOutput(awsUploader)
			awsDownloader = awsTest.NewDownloader()
			awsS3Manager.SetDownloaderOutput(awsDownloader)
		})

		AfterEach(func() {
			awsDownloader.AssertOutputsEmpty()
			awsUploader.AssertOutputsEmpty()
			awsS3Manager.AssertOutputsEmpty()
			awsS3.AssertOutputsEmpty()
			awsAPI.AssertOutputsEmpty()
		})

		Context("NewStore", func() {
			It("return an error if the config is missing", func() {
				store, err := storeUnstructuredS3.NewStore(nil, awsAPI)
				Expect(err).To(MatchError("config is missing"))
				Expect(store).To(BeNil())
			})

			It("return an error if the config is invalid", func() {
				config.Bucket = ""
				store, err := storeUnstructuredS3.NewStore(config, awsAPI)
				Expect(err).To(MatchError("config is invalid; bucket is missing"))
				Expect(store).To(BeNil())
			})

			It("return an error if the aws api is missing", func() {
				store, err := storeUnstructuredS3.NewStore(config, nil)
				Expect(err).To(MatchError("aws api is missing"))
				Expect(store).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(storeUnstructuredS3.NewStore(config, awsAPI)).ToNot(BeNil())
			})
		})

		Context("with new store", func() {
			var store *storeUnstructuredS3.Store
			var ctx context.Context
			var key string
			var keyPath string
			var contents []byte

			BeforeEach(func() {
				var err error
				store, err = storeUnstructuredS3.NewStore(config, awsAPI)
				Expect(err).ToNot(HaveOccurred())
				Expect(store).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				key = storeUnstructuredTest.RandomKey()
				keyPath = fmt.Sprintf("%s/%s", config.Prefix, key)
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

				Context("with aws s3 head object", func() {
					AfterEach(func() {
						Expect(awsS3.HeadObjectWithContextInputs).To(HaveLen(1))
						Expect(awsS3.HeadObjectWithContextInputs[0].Input).To(Equal(&s3.HeadObjectInput{
							Bucket: pointer.FromString(config.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsS3.HeadObjectWithContextInputs[0].Options).To(BeEmpty())
					})

					It("returns an error if aws returns an error, but not awserr.Error", func() {
						awsErr := errorsTest.RandomError()
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Exists(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns an error if aws returns an awserr.Error, but not NotFound", func() {
						awsErr := awserr.New(test.RandomString(), "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Exists(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns false if the key does not exist", func() {
						awsErr := awserr.New("NotFound", "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Exists(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(BeFalse())
					})

					It("returns true if the key exists", func() {
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: nil}}
						exists, err := store.Exists(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(BeTrue())
					})
				})
			})

			Context("Put", func() {
				var reader io.Reader

				BeforeEach(func() {
					reader = bytes.NewReader(contents)
				})

				It("returns an error if the context is missing", func() {
					Expect(store.Put(nil, key, reader)).To(MatchError("context is missing"))
				})

				It("returns an error if the key is missing", func() {
					Expect(store.Put(ctx, "", reader)).To(MatchError("key is missing"))
				})

				It("returns an error if the key is invalid", func() {
					Expect(store.Put(ctx, "#invalid#", reader)).To(MatchError("key is invalid"))
				})

				It("returns an error if the reader is missing", func() {
					Expect(store.Put(ctx, key, nil)).To(MatchError("reader is missing"))
				})

				Context("with aws s3 manager upload", func() {
					AfterEach(func() {
						Expect(awsUploader.UploadWithContextInputs).To(HaveLen(1))
						Expect(awsUploader.UploadWithContextInputs[0].Options).To(BeEmpty())
						awsUploader.AssertOutputsEmpty()
					})

					Context("without options", func() {
						AfterEach(func() {
							Expect(awsUploader.UploadWithContextInputs[0].Input).To(Equal(&s3manager.UploadInput{
								Body:                 reader,
								Bucket:               pointer.FromString(config.Bucket),
								Key:                  pointer.FromString(keyPath),
								ServerSideEncryption: pointer.FromString("AES256"),
							}))
						})

						It("returns an error if aws returns an error", func() {
							awsErr := errorsTest.RandomError()
							awsUploader.UploadWithContextOutputs = []awsTest.UploadWithContextOutput{{Output: nil, Error: awsErr}}
							Expect(store.Put(ctx, key, reader)).To(MatchError(fmt.Sprintf("unable to upload object with key %q; %s", keyPath, awsErr)))
						})

						It("returns successfully", func() {
							awsUploader.UploadWithContextOutputs = []awsTest.UploadWithContextOutput{{Output: nil, Error: nil}}
							Expect(store.Put(ctx, key, reader)).To(Succeed())
						})
					})
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

				Context("with aws s3 manager download", func() {
					AfterEach(func() {
						Expect(awsDownloader.DownloadWithContextInputs).To(HaveLen(1))
						Expect(awsDownloader.DownloadWithContextInputs[0].WriterAt).ToNot(BeNil())
						Expect(awsDownloader.DownloadWithContextInputs[0].Input).To(Equal(&s3.GetObjectInput{
							Bucket: pointer.FromString(config.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsDownloader.DownloadWithContextInputs[0].Options).To(BeEmpty())
						awsDownloader.AssertOutputsEmpty()
					})

					It("returns an error if aws returns an error, but not awserr.Error", func() {
						awsErr := errorsTest.RandomError()
						awsDownloader.DownloadWithContextOutputs = []awsTest.DownloadWithContextOutput{{BytesWritten: 0, Error: awsErr}}
						var err error
						reader, err = store.Get(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to download object with key %q; %s", keyPath, awsErr)))
						Expect(reader).To(BeNil())
					})

					It("returns an error if aws returns an awserr.Error, but not NoSuchKey", func() {
						awsErr := awserr.New(test.RandomString(), "", nil)
						awsDownloader.DownloadWithContextOutputs = []awsTest.DownloadWithContextOutput{{BytesWritten: 0, Error: awsErr}}
						var err error
						reader, err = store.Get(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to download object with key %q; %s", keyPath, awsErr)))
						Expect(reader).To(BeNil())
					})

					It("returns nil if the key does not exist", func() {
						awsErr := awserr.New("NoSuchKey", "", nil)
						awsDownloader.DownloadWithContextOutputs = []awsTest.DownloadWithContextOutput{{BytesWritten: 0, Error: awsErr}}
						var err error
						reader, err = store.Get(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader).To(BeNil())
					})

					It("returns reader if the key exists", func() {
						awsDownloader.DownloadWithContextStub = func(ctx awsSdkAws.Context, writerAt io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (int64, error) {
							Expect(writerAt.WriteAt(contents, 0)).To(Equal(len(contents)))
							return 0, nil
						}
						var err error
						reader, err = store.Get(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader).ToNot(BeNil())
						Expect(ioutil.ReadAll(reader)).To(Equal(contents))
					})
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

				Context("with aws s3 head object", func() {
					AfterEach(func() {
						Expect(awsS3.HeadObjectWithContextInputs).To(HaveLen(1))
						Expect(awsS3.HeadObjectWithContextInputs[0].Input).To(Equal(&s3.HeadObjectInput{
							Bucket: pointer.FromString(config.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsS3.HeadObjectWithContextInputs[0].Options).To(BeEmpty())
					})

					It("returns an error if aws returns an error, but not awserr.Error", func() {
						awsErr := errorsTest.RandomError()
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns an error if aws returns an awserr.Error, but not NotFound", func() {
						awsErr := awserr.New(test.RandomString(), "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns false if the key does not exist", func() {
						awsErr := awserr.New("NotFound", "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(BeFalse())
					})

					Context("with aws s3 delete object", func() {
						BeforeEach(func() {
							awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: nil}}
						})

						AfterEach(func() {
							Expect(awsS3.DeleteObjectWithContextInputs).To(HaveLen(1))
							Expect(awsS3.DeleteObjectWithContextInputs[0].Input).To(Equal(&s3.DeleteObjectInput{
								Bucket: pointer.FromString(config.Bucket),
								Key:    pointer.FromString(keyPath),
							}))
							Expect(awsS3.DeleteObjectWithContextInputs[0].Options).To(BeEmpty())
						})

						It("returns an error if aws returns an error", func() {
							awsErr := errorsTest.RandomError()
							awsS3.DeleteObjectWithContextOutputs = []awsTest.DeleteObjectWithContextOutput{{Output: nil, Error: awsErr}}
							exists, err := store.Delete(ctx, key)
							Expect(err).To(MatchError(fmt.Sprintf("unable to delete object with key %q; %s", keyPath, awsErr)))
							Expect(exists).To(BeFalse())
						})

						It("returns true if the key exists and it deletes the file", func() {
							awsS3.DeleteObjectWithContextOutputs = []awsTest.DeleteObjectWithContextOutput{{Output: nil, Error: nil}}
							Expect(store.Delete(ctx, key)).To(BeTrue())
						})
					})
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

				Context("with aws s3 head object", func() {
					AfterEach(func() {
						Expect(awsS3.HeadObjectWithContextInputs).To(HaveLen(1))
						Expect(awsS3.HeadObjectWithContextInputs[0].Input).To(Equal(&s3.HeadObjectInput{
							Bucket: pointer.FromString(config.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsS3.HeadObjectWithContextInputs[0].Options).To(BeEmpty())
					})

					It("returns an error if aws returns an error, but not awserr.Error", func() {
						awsErr := errorsTest.RandomError()
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns an error if aws returns an awserr.Error, but not NotFound", func() {
						awsErr := awserr.New(test.RandomString(), "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns false if the key does not exist", func() {
						awsErr := awserr.New("NotFound", "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(BeFalse())
					})

					Context("with aws s3 delete object", func() {
						BeforeEach(func() {
							awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: nil}}
						})

						AfterEach(func() {
							Expect(awsS3.DeleteObjectWithContextInputs).To(HaveLen(1))
							Expect(awsS3.DeleteObjectWithContextInputs[0].Input).To(Equal(&s3.DeleteObjectInput{
								Bucket: pointer.FromString(config.Bucket),
								Key:    pointer.FromString(keyPath),
							}))
							Expect(awsS3.DeleteObjectWithContextInputs[0].Options).To(BeEmpty())
						})

						It("returns an error if aws returns an error", func() {
							awsErr := errorsTest.RandomError()
							awsS3.DeleteObjectWithContextOutputs = []awsTest.DeleteObjectWithContextOutput{{Output: nil, Error: awsErr}}
							exists, err := store.Delete(ctx, key)
							Expect(err).To(MatchError(fmt.Sprintf("unable to delete object with key %q; %s", keyPath, awsErr)))
							Expect(exists).To(BeFalse())
						})

						It("returns true if the key exists and it deletes the file", func() {
							awsS3.DeleteObjectWithContextOutputs = []awsTest.DeleteObjectWithContextOutput{{Output: nil, Error: nil}}
							Expect(store.Delete(ctx, key)).To(BeTrue())
						})
					})
				})
			})
		})
	})
})
