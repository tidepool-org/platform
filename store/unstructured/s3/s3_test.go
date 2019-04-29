package s3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"

	awsTest "github.com/tidepool-org/platform/aws/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
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
		var cfg *storeUnstructuredS3.Config
		var awsAPI *awsTest.API

		BeforeEach(func() {
			cfg = storeUnstructuredS3.NewConfig()
			Expect(cfg).ToNot(BeNil())
			cfg.Bucket = awsTest.RandomBucket()
			cfg.Prefix = storeUnstructuredTest.RandomKey()
			awsAPI = awsTest.NewAPI()
		})

		AfterEach(func() {
			awsAPI.AssertOutputsEmpty()
		})

		Context("NewStore", func() {
			It("return an error if the config is missing", func() {
				str, err := storeUnstructuredS3.NewStore(nil, awsAPI)
				Expect(err).To(MatchError("config is missing"))
				Expect(str).To(BeNil())
			})

			It("return an error if the config is invalid", func() {
				cfg.Bucket = ""
				str, err := storeUnstructuredS3.NewStore(cfg, awsAPI)
				Expect(err).To(MatchError("config is invalid; bucket is missing"))
				Expect(str).To(BeNil())
			})

			It("return an error if the aws api is missing", func() {
				str, err := storeUnstructuredS3.NewStore(cfg, nil)
				Expect(err).To(MatchError("aws api is missing"))
				Expect(str).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(storeUnstructuredS3.NewStore(cfg, awsAPI)).ToNot(BeNil())
			})
		})

		Context("with new store", func() {
			var str *storeUnstructuredS3.Store
			var ctx context.Context
			var key string
			var keyPath string
			var contents []byte

			BeforeEach(func() {
				var err error
				str, err = storeUnstructuredS3.NewStore(cfg, awsAPI)
				Expect(err).ToNot(HaveOccurred())
				Expect(str).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
				key = storeUnstructuredTest.RandomKey()
				keyPath = fmt.Sprintf("%s/%s", cfg.Prefix, key)
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

				Context("with aws s3 head object", func() {
					var awsS3 *awsTest.S3

					BeforeEach(func() {
						awsS3 = awsTest.NewS3()
						awsAPI.S3Outputs = []s3iface.S3API{awsS3}
					})

					AfterEach(func() {
						Expect(awsS3.HeadObjectWithContextInputs).To(HaveLen(1))
						Expect(awsS3.HeadObjectWithContextInputs[0].Context).To(Equal(ctx))
						Expect(awsS3.HeadObjectWithContextInputs[0].Input).To(Equal(&s3.HeadObjectInput{
							Bucket: pointer.FromString(cfg.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsS3.HeadObjectWithContextInputs[0].Options).To(BeEmpty())
						awsS3.AssertOutputsEmpty()
					})

					It("returns an error if aws returns an error, but not awserr.Error", func() {
						awsErr := errorsTest.RandomError()
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := str.Exists(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns an error if aws returns an awserr.Error, but not NotFound", func() {
						awsErr := awserr.New(test.RandomString(), "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := str.Exists(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns false if the key does not exist", func() {
						awsErr := awserr.New("NotFound", "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := str.Exists(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(BeFalse())
					})

					It("returns true if the key exists", func() {
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: nil}}
						exists, err := str.Exists(ctx, key)
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

				Context("with aws s3 manager upload", func() {
					var awsS3Manager *awsTest.S3Manager

					BeforeEach(func() {
						awsS3Manager = awsTest.NewS3Manager()
						awsAPI.S3ManagerUploaderOutputs = []s3manageriface.UploaderAPI{awsS3Manager}
					})

					AfterEach(func() {
						Expect(awsS3Manager.UploadWithContextInputs).To(HaveLen(1))
						Expect(awsS3Manager.UploadWithContextInputs[0].Context).To(Equal(ctx))
						Expect(awsS3Manager.UploadWithContextInputs[0].Input).To(Equal(&s3manager.UploadInput{
							Body:                 reader,
							Bucket:               pointer.FromString(cfg.Bucket),
							Key:                  pointer.FromString(keyPath),
							ServerSideEncryption: pointer.FromString("AES256"),
						}))
						Expect(awsS3Manager.UploadWithContextInputs[0].Options).To(BeEmpty())
						awsS3Manager.AssertOutputsEmpty()
					})

					It("returns an error if aws returns an error", func() {
						awsErr := errorsTest.RandomError()
						awsS3Manager.UploadWithContextOutputs = []awsTest.UploadWithContextOutput{{Output: nil, Error: awsErr}}
						Expect(str.Put(ctx, key, reader)).To(MatchError(fmt.Sprintf("unable to upload object with key %q; %s", keyPath, awsErr)))
					})

					It("returns successfully", func() {
						awsS3Manager.UploadWithContextOutputs = []awsTest.UploadWithContextOutput{{Output: nil, Error: nil}}
						Expect(str.Put(ctx, key, reader)).To(Succeed())
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

				Context("with aws s3 manager download", func() {
					var awsS3Manager *awsTest.S3Manager

					BeforeEach(func() {
						awsS3Manager = awsTest.NewS3Manager()
						awsAPI.S3ManagerDownloaderOutputs = []s3manageriface.DownloaderAPI{awsS3Manager}
					})

					AfterEach(func() {
						Expect(awsS3Manager.DownloadWithContextInputs).To(HaveLen(1))
						Expect(awsS3Manager.DownloadWithContextInputs[0].Context).To(Equal(ctx))
						Expect(awsS3Manager.DownloadWithContextInputs[0].WriterAt).ToNot(BeNil())
						Expect(awsS3Manager.DownloadWithContextInputs[0].Input).To(Equal(&s3.GetObjectInput{
							Bucket: pointer.FromString(cfg.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsS3Manager.DownloadWithContextInputs[0].Options).To(BeEmpty())
						awsS3Manager.AssertOutputsEmpty()
					})

					It("returns an error if aws returns an error, but not awserr.Error", func() {
						awsErr := errorsTest.RandomError()
						awsS3Manager.DownloadWithContextOutputs = []awsTest.DownloadWithContextOutput{{BytesWritten: 0, Error: awsErr}}
						var err error
						reader, err = str.Get(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to download object with key %q; %s", keyPath, awsErr)))
						Expect(reader).To(BeNil())
					})

					It("returns an error if aws returns an awserr.Error, but not NoSuchKey", func() {
						awsErr := awserr.New(test.RandomString(), "", nil)
						awsS3Manager.DownloadWithContextOutputs = []awsTest.DownloadWithContextOutput{{BytesWritten: 0, Error: awsErr}}
						var err error
						reader, err = str.Get(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to download object with key %q; %s", keyPath, awsErr)))
						Expect(reader).To(BeNil())
					})

					It("returns nil if the key does not exist", func() {
						awsErr := awserr.New("NoSuchKey", "", nil)
						awsS3Manager.DownloadWithContextOutputs = []awsTest.DownloadWithContextOutput{{BytesWritten: 0, Error: awsErr}}
						var err error
						reader, err = str.Get(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader).To(BeNil())
					})

					It("returns reader if the key exists", func() {
						awsS3Manager.DownloadWithContextStub = func(ctx aws.Context, writerAt io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (int64, error) {
							Expect(writerAt.WriteAt(contents, 0)).To(Equal(len(contents)))
							return 0, nil
						}
						var err error
						reader, err = str.Get(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(reader).ToNot(BeNil())
						Expect(ioutil.ReadAll(reader)).To(Equal(contents))
					})
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

				Context("with aws s3 head object", func() {
					var awsS3 *awsTest.S3

					BeforeEach(func() {
						awsS3 = awsTest.NewS3()
						awsAPI.S3Outputs = []s3iface.S3API{awsS3}
					})

					AfterEach(func() {
						Expect(awsS3.HeadObjectWithContextInputs).To(HaveLen(1))
						Expect(awsS3.HeadObjectWithContextInputs[0].Context).To(Equal(ctx))
						Expect(awsS3.HeadObjectWithContextInputs[0].Input).To(Equal(&s3.HeadObjectInput{
							Bucket: pointer.FromString(cfg.Bucket),
							Key:    pointer.FromString(keyPath),
						}))
						Expect(awsS3.HeadObjectWithContextInputs[0].Options).To(BeEmpty())
						awsS3.AssertOutputsEmpty()
					})

					It("returns an error if aws returns an error, but not awserr.Error", func() {
						awsErr := errorsTest.RandomError()
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := str.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns an error if aws returns an awserr.Error, but not NotFound", func() {
						awsErr := awserr.New(test.RandomString(), "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := str.Delete(ctx, key)
						Expect(err).To(MatchError(fmt.Sprintf("unable to head object with key %q; %s", keyPath, awsErr)))
						Expect(exists).To(BeFalse())
					})

					It("returns false if the key does not exist", func() {
						awsErr := awserr.New("NotFound", "", nil)
						awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: awsErr}}
						exists, err := str.Delete(ctx, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(BeFalse())
					})

					Context("with aws s3 delete object", func() {
						BeforeEach(func() {
							awsS3.HeadObjectWithContextOutputs = []awsTest.HeadObjectWithContextOutput{{Output: nil, Error: nil}}
							awsAPI.S3Outputs = append(awsAPI.S3Outputs, awsS3)
						})

						AfterEach(func() {
							Expect(awsS3.DeleteObjectWithContextInputs).To(HaveLen(1))
							Expect(awsS3.DeleteObjectWithContextInputs[0].Context).To(Equal(ctx))
							Expect(awsS3.DeleteObjectWithContextInputs[0].Input).To(Equal(&s3.DeleteObjectInput{
								Bucket: pointer.FromString(cfg.Bucket),
								Key:    pointer.FromString(keyPath),
							}))
							Expect(awsS3.DeleteObjectWithContextInputs[0].Options).To(BeEmpty())
						})

						It("returns an error if aws returns an error", func() {
							awsErr := errorsTest.RandomError()
							awsS3.DeleteObjectWithContextOutputs = []awsTest.DeleteObjectWithContextOutput{{Output: nil, Error: awsErr}}
							exists, err := str.Delete(ctx, key)
							Expect(err).To(MatchError(fmt.Sprintf("unable to delete object with key %q; %s", keyPath, awsErr)))
							Expect(exists).To(BeFalse())
						})

						It("returns true if the key exists and it deletes the file", func() {
							awsS3.DeleteObjectWithContextOutputs = []awsTest.DeleteObjectWithContextOutput{{Output: nil, Error: nil}}
							Expect(str.Delete(ctx, key)).To(BeTrue())
						})
					})
				})
			})
		})
	})
})
