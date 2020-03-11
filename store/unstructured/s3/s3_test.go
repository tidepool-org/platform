package s3_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	awsSdkGoAws "github.com/aws/aws-sdk-go/aws"
	awsSdkGoAwsAwserr "github.com/aws/aws-sdk-go/aws/awserr"
	awsSdkGoServiceS3 "github.com/aws/aws-sdk-go/service/s3"
	awsSdkGoServiceS3S3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/aws"
	awsTest "github.com/tidepool-org/platform/aws/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
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
						Expect(awsS3.HeadObjectWithContextInputs[0].Input).To(Equal(&awsSdkGoServiceS3.HeadObjectInput{
							Bucket: pointer.FromString(config.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsS3.HeadObjectWithContextInputs[0].Options).To(BeEmpty())
					})

					It("returns an error if aws returns an error, but not Error", func() {
						awsErr := errorsTest.RandomError()
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Exists(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns an error if aws returns an Error, but not NotFound", func() {
						awsErr := awsSdkGoAwsAwserr.New(test.RandomString(), "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Exists(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns false if the key does not exist", func() {
						awsErr := awsSdkGoAwsAwserr.New("NotFound", "", nil)
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

				Context("with aws s3 manager upload", func() {
					AfterEach(func() {
						Expect(awsUploader.UploadWithContextInputs).To(HaveLen(1))
						Expect(awsUploader.UploadWithContextInputs[0].Options).To(BeEmpty())
						awsUploader.AssertOutputsEmpty()
					})

					Context("without options", func() {
						BeforeEach(func() {
							options = nil
						})

						AfterEach(func() {
							Expect(awsUploader.UploadWithContextInputs[0].Input).To(Equal(&awsSdkGoServiceS3S3manager.UploadInput{
								Body:                 reader,
								Bucket:               pointer.FromString(config.Bucket),
								Key:                  pointer.FromString(keyPath),
								ServerSideEncryption: pointer.FromString("AES256"),
							}))
						})

						It("returns an error if aws returns an error", func() {
							awsErr := errorsTest.RandomError()
							awsUploader.UploadWithContextOutputs = []awsTest.UploadWithContextOutput{{Output: nil, Error: awsErr}}
							Expect(store.Put(ctx, key, reader, options)).To(MatchError(fmt.Sprintf("unable to upload object with key %q, bucket %q; %s,", keyPath, config.Bucket, awsErr)))
						})

						It("returns successfully", func() {
							awsUploader.UploadWithContextOutputs = []awsTest.UploadWithContextOutput{{Output: nil, Error: nil}}
							Expect(store.Put(ctx, key, reader, options)).To(Succeed())
						})
					})

					Context("with options", func() {
						AfterEach(func() {
							Expect(awsUploader.UploadWithContextInputs[0].Input).To(Equal(&awsSdkGoServiceS3S3manager.UploadInput{
								Body:                 reader,
								Bucket:               pointer.FromString(config.Bucket),
								ContentType:          options.MediaType,
								Key:                  pointer.FromString(keyPath),
								ServerSideEncryption: pointer.FromString("AES256"),
							}))
						})

						It("returns an error if aws returns an error", func() {
							awsErr := errorsTest.RandomError()
							awsUploader.UploadWithContextOutputs = []awsTest.UploadWithContextOutput{{Output: nil, Error: awsErr}}
							Expect(store.Put(ctx, key, reader, options)).To(MatchError(fmt.Sprintf("unable to upload object with key %q, bucket %q;; %s", keyPath, config.Bucket, awsErr)))
						})

						It("returns successfully", func() {
							awsUploader.UploadWithContextOutputs = []awsTest.UploadWithContextOutput{{Output: nil, Error: nil}}
							Expect(store.Put(ctx, key, reader, options)).To(Succeed())
						})

						It("returns successfully without options media type", func() {
							options.MediaType = nil
							awsUploader.UploadWithContextOutputs = []awsTest.UploadWithContextOutput{{Output: nil, Error: nil}}
							Expect(store.Put(ctx, key, reader, options)).To(Succeed())
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
						Expect(awsDownloader.DownloadWithContextInputs[0].Input).To(Equal(&awsSdkGoServiceS3.GetObjectInput{
							Bucket: pointer.FromString(config.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsDownloader.DownloadWithContextInputs[0].Options).To(BeEmpty())
						awsDownloader.AssertOutputsEmpty()
					})

					It("returns an error if aws returns an error, but not Error", func() {
						awsErr := errorsTest.RandomError()
						awsDownloader.DownloadWithContextOutputs = []awsTest.DownloadWithContextOutput{{BytesWritten: 0, Error: awsErr}}
						var err error
						reader, err = store.Get(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to download object with key %q; %s", keyPath, awsErr)))
						Expect(reader).To(BeNil())
					})

					It("returns an error if aws returns an Error, but not NoSuchKey", func() {
						awsErr := awsSdkGoAwsAwserr.New(test.RandomString(), "", nil)
						awsDownloader.DownloadWithContextOutputs = []awsTest.DownloadWithContextOutput{{BytesWritten: 0, Error: awsErr}}
						var err error
						reader, err = store.Get(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to download object with key %q; %s", keyPath, awsErr)))
						Expect(reader).To(BeNil())
					})

					It("returns nil if the key does not exist", func() {
						awsErr := awsSdkGoAwsAwserr.New("NoSuchKey", "", nil)
						awsDownloader.DownloadWithContextOutputs = []awsTest.DownloadWithContextOutput{{BytesWritten: 0, Error: awsErr}}
						var err error
						reader, err = store.Get(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader).To(BeNil())
					})

					It("returns reader if the key exists", func() {
						awsDownloader.DownloadWithContextStub = func(ctx awsSdkGoAws.Context, writerAt io.WriterAt, input *awsSdkGoServiceS3.GetObjectInput, options ...func(*awsSdkGoServiceS3S3manager.Downloader)) (int64, error) {
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
						Expect(awsS3.HeadObjectWithContextInputs[0].Input).To(Equal(&awsSdkGoServiceS3.HeadObjectInput{
							Bucket: pointer.FromString(config.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsS3.HeadObjectWithContextInputs[0].Options).To(BeEmpty())
					})

					It("returns an error if aws returns an error, but not Error", func() {
						awsErr := errorsTest.RandomError()
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns an error if aws returns an Error, but not NotFound", func() {
						awsErr := awsSdkGoAwsAwserr.New(test.RandomString(), "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns false if the key does not exist", func() {
						awsErr := awsSdkGoAwsAwserr.New("NotFound", "", nil)
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
							Expect(awsS3.DeleteObjectWithContextInputs[0].Input).To(Equal(&awsSdkGoServiceS3.DeleteObjectInput{
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
						Expect(awsS3.HeadObjectWithContextInputs[0].Input).To(Equal(&awsSdkGoServiceS3.HeadObjectInput{
							Bucket: pointer.FromString(config.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsS3.HeadObjectWithContextInputs[0].Options).To(BeEmpty())
					})

					It("returns an error if aws returns an error, but not Error", func() {
						awsErr := errorsTest.RandomError()
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns an error if aws returns an Error, but not NotFound", func() {
						awsErr := awsSdkGoAwsAwserr.New(test.RandomString(), "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := store.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns false if the key does not exist", func() {
						awsErr := awsSdkGoAwsAwserr.New("NotFound", "", nil)
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
							Expect(awsS3.DeleteObjectWithContextInputs[0].Input).To(Equal(&awsSdkGoServiceS3.DeleteObjectInput{
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

				Context("with batch delete with client and delete list iterator", func() {
					var awsBatchDeleteWithClient *awsTest.BatchDeleteWithClient
					var awsBatchDeleteIterator *awsTest.BatchDeleteIterator

					BeforeEach(func() {
						awsBatchDeleteWithClient = awsTest.NewBatchDeleteWithClient()
						awsS3Manager.NewBatchDeleteWithClientOutputs = []aws.BatchDeleteWithClient{awsBatchDeleteWithClient}
						awsBatchDeleteIterator = awsTest.NewBatchDeleteIterator()
						awsS3Manager.NewDeleteListIteratorOutputs = []awsSdkGoServiceS3S3manager.BatchDeleteIterator{awsBatchDeleteIterator}
					})

					AfterEach(func() {
						Expect(awsS3Manager.NewBatchDeleteWithClientInputs).To(HaveLen(1))
						Expect(awsS3Manager.NewBatchDeleteWithClientInputs[0]).ToNot(BeNil())
						batchDelete := &awsSdkGoServiceS3S3manager.BatchDelete{}
						for _, option := range awsS3Manager.NewBatchDeleteWithClientInputs[0] {
							option(batchDelete)
						}
						Expect(batchDelete.BatchSize).To(Equal(1000))
						Expect(awsS3Manager.NewDeleteListIteratorInputs).To(HaveLen(1))
						Expect(awsS3Manager.NewDeleteListIteratorInputs[0]).ToNot(BeNil())
						Expect(awsS3Manager.NewDeleteListIteratorInputs[0].ListObjectsInput).ToNot(BeNil())
						Expect(awsS3Manager.NewDeleteListIteratorInputs[0].ListObjectsInput.Bucket).To(Equal(pointer.FromString(config.Bucket)))
						Expect(awsS3Manager.NewDeleteListIteratorInputs[0].ListObjectsInput.Prefix).To(Equal(pointer.FromString(keyPath)))
						Expect(awsS3Manager.NewDeleteListIteratorInputs[0].ListObjectsInput.MaxKeys).ToNot(BeNil())
						Expect(*awsS3Manager.NewDeleteListIteratorInputs[0].ListObjectsInput.MaxKeys).To(Equal(int64(1000)))
						Expect(awsS3Manager.NewDeleteListIteratorInputs[0].Options).To(BeEmpty())
						Expect(awsBatchDeleteWithClient.DeleteInputs).To(Equal([]awsSdkGoServiceS3S3manager.BatchDeleteIterator{awsBatchDeleteIterator}))
						awsBatchDeleteWithClient.AssertOutputsEmpty()
					})

					It("returns an error if batch delete returns an error", func() {
						awsErr := errorsTest.RandomError()
						awsBatchDeleteWithClient.DeleteOutputs = []error{awsErr}
						errorsTest.ExpectEqual(store.DeleteDirectory(ctx, key), errors.Wrapf(awsErr, "unable to delete all objects with key %q", keyPath))
					})

					It("returns successfully if batch delete returns successfully", func() {
						awsBatchDeleteWithClient.DeleteOutputs = []error{nil}
						Expect(store.DeleteDirectory(ctx, key)).To(Succeed())
					})
				})
			})
		})
	})
})
