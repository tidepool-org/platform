package unstructured_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	imageStoreUnstructured "github.com/tidepool-org/platform/image/store/unstructured"
	imageTest "github.com/tidepool-org/platform/image/test"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
	storeUnstructuredTest "github.com/tidepool-org/platform/store/unstructured/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Unstructured", func() {
	var underlyingStore *storeUnstructuredTest.Store

	BeforeEach(func() {
		underlyingStore = storeUnstructuredTest.NewStore()
	})

	AfterEach(func() {
		underlyingStore.AssertOutputsEmpty()
	})

	Context("NewStore", func() {
		It("returns an error when the store is missing", func() {
			store, err := imageStoreUnstructured.NewStore(nil)
			errorsTest.ExpectEqual(err, errors.New("store is missing"))
			Expect(store).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(imageStoreUnstructured.NewStore(underlyingStore)).ToNot(BeNil())
		})
	})

	Context("with new store, user id, and image id", func() {
		var store *imageStoreUnstructured.StoreImpl
		var ctx context.Context
		var userID string
		var imageID string
		var key string

		BeforeEach(func() {
			var err error
			store, err = imageStoreUnstructured.NewStore(underlyingStore)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			ctx = context.Background()
			userID = userTest.RandomID()
			imageID = imageTest.RandomID()
		})

		Context("with content id", func() {
			var contentID string

			BeforeEach(func() {
				contentID = imageTest.RandomContentID()
			})

			Context("with content intent", func() {
				var contentIntent string

				BeforeEach(func() {
					contentIntent = imageTest.RandomContentIntent()
					key = fmt.Sprintf("%s/%s/content/%s/%s", userID, imageID, contentID, contentIntent)
				})

				Context("PutContent", func() {
					var reader io.Reader
					var options *storeUnstructured.Options

					BeforeEach(func() {
						reader = bytes.NewReader(imageTest.RandomContentBytes())
						options = storeUnstructuredTest.RandomOptions()
					})

					AfterEach(func() {
						Expect(underlyingStore.PutInputs).To(Equal([]storeUnstructuredTest.PutInput{{Key: key, Reader: reader, Options: options}}))
					})

					It("returns an error when the underlying store returns an error", func() {
						parentErr := errorsTest.RandomError()
						underlyingStore.PutOutputs = []error{parentErr}
						errorsTest.ExpectEqual(store.PutContent(ctx, userID, imageID, contentID, contentIntent, reader, options), errors.Wrap(parentErr, "unable to put image content"))
					})

					It("returns successfully when the underlying store returns successfully", func() {
						underlyingStore.PutOutputs = []error{nil}
						Expect(store.PutContent(ctx, userID, imageID, contentID, contentIntent, reader, options)).To(Succeed())
					})
				})

				Context("GetContent", func() {
					var parentReader io.ReadCloser

					BeforeEach(func() {
						parentReader = ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
					})

					AfterEach(func() {
						Expect(underlyingStore.GetInputs).To(Equal([]string{key}))
					})

					It("returns an error when the underlying store returns an error", func() {
						parentErr := errorsTest.RandomError()
						underlyingStore.GetOutputs = []storeUnstructuredTest.GetOutput{{Reader: nil, Error: parentErr}}
						reader, err := store.GetContent(ctx, userID, imageID, contentID, contentIntent)
						errorsTest.ExpectEqual(err, errors.Wrap(parentErr, "unable to get image content"))
						Expect(reader).To(BeNil())
					})

					It("returns nil when the underlying store returns nil", func() {
						underlyingStore.GetOutputs = []storeUnstructuredTest.GetOutput{{Reader: nil, Error: nil}}
						Expect(store.GetContent(ctx, userID, imageID, contentID, contentIntent)).To(BeNil())
					})

					It("returns a reader when the underlying store returns a reader", func() {
						underlyingStore.GetOutputs = []storeUnstructuredTest.GetOutput{{Reader: parentReader, Error: nil}}
						Expect(store.GetContent(ctx, userID, imageID, contentID, contentIntent)).To(Equal(parentReader))
					})
				})
			})

			Context("DeleteContent", func() {
				BeforeEach(func() {
					key = fmt.Sprintf("%s/%s/content/%s", userID, imageID, contentID)
				})

				AfterEach(func() {
					Expect(underlyingStore.DeleteDirectoryInputs).To(Equal([]string{key}))
				})

				It("returns an error when the underlying store returns an error", func() {
					parentErr := errorsTest.RandomError()
					underlyingStore.DeleteDirectoryOutputs = []error{parentErr}
					errorsTest.ExpectEqual(store.DeleteContent(ctx, userID, imageID, contentID), errors.Wrap(parentErr, "unable to delete all image content"))
				})

				It("returns successfully when the underlying store returns successfully", func() {
					underlyingStore.DeleteDirectoryOutputs = []error{nil}
					Expect(store.DeleteContent(ctx, userID, imageID, contentID)).To(Succeed())
				})
			})

			Context("with rendition id", func() {
				var renditionsID string

				BeforeEach(func() {
					renditionsID = imageTest.RandomRenditionsID()
				})

				Context("rendition", func() {
					var rendition string

					BeforeEach(func() {
						rendition = imageTest.RandomRenditionAsString()
						key = fmt.Sprintf("%s/%s/content/%s/renditions/%s/%s", userID, imageID, contentID, renditionsID, rendition)
					})

					Context("PutRenditionContent", func() {
						var reader io.Reader
						var options *storeUnstructured.Options

						BeforeEach(func() {
							reader = bytes.NewReader(imageTest.RandomContentBytes())
							options = storeUnstructuredTest.RandomOptions()
						})

						AfterEach(func() {
							Expect(underlyingStore.PutInputs).To(Equal([]storeUnstructuredTest.PutInput{{Key: key, Reader: reader, Options: options}}))
						})

						It("returns an error when the underlying store returns an error", func() {
							parentErr := errorsTest.RandomError()
							underlyingStore.PutOutputs = []error{parentErr}
							errorsTest.ExpectEqual(store.PutRenditionContent(ctx, userID, imageID, contentID, renditionsID, rendition, reader, options), errors.Wrap(parentErr, "unable to put image rendition content"))
						})

						It("returns successfully when the underlying store returns successfully", func() {
							underlyingStore.PutOutputs = []error{nil}
							Expect(store.PutRenditionContent(ctx, userID, imageID, contentID, renditionsID, rendition, reader, options)).To(Succeed())
						})
					})

					Context("GetRenditionContent", func() {
						var parentReader io.ReadCloser

						BeforeEach(func() {
							parentReader = ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
						})

						AfterEach(func() {
							Expect(underlyingStore.GetInputs).To(Equal([]string{key}))
						})

						It("returns an error when the underlying store returns an error", func() {
							parentErr := errorsTest.RandomError()
							underlyingStore.GetOutputs = []storeUnstructuredTest.GetOutput{{Reader: nil, Error: parentErr}}
							reader, err := store.GetRenditionContent(ctx, userID, imageID, contentID, renditionsID, rendition)
							errorsTest.ExpectEqual(err, errors.Wrap(parentErr, "unable to get image rendition content"))
							Expect(reader).To(BeNil())
						})

						It("returns nil when the underlying store returns nil", func() {
							underlyingStore.GetOutputs = []storeUnstructuredTest.GetOutput{{Reader: nil, Error: nil}}
							Expect(store.GetRenditionContent(ctx, userID, imageID, contentID, renditionsID, rendition)).To(BeNil())
						})

						It("returns a reader when the underlying store returns a reader", func() {
							underlyingStore.GetOutputs = []storeUnstructuredTest.GetOutput{{Reader: parentReader, Error: nil}}
							Expect(store.GetRenditionContent(ctx, userID, imageID, contentID, renditionsID, rendition)).To(Equal(parentReader))
						})
					})
				})

				Context("DeleteRenditionContent", func() {
					AfterEach(func() {
						Expect(underlyingStore.DeleteDirectoryInputs).To(Equal([]string{fmt.Sprintf("%s/%s/content/%s/renditions/%s", userID, imageID, contentID, renditionsID)}))
					})

					It("returns an error when the underlying store returns an error", func() {
						parentErr := errorsTest.RandomError()
						underlyingStore.DeleteDirectoryOutputs = []error{parentErr}
						errorsTest.ExpectEqual(store.DeleteRenditionContent(ctx, userID, imageID, contentID, renditionsID), errors.Wrap(parentErr, "unable to delete image rendition content"))
					})

					It("returns successfully when the underlying store returns successfully", func() {
						underlyingStore.DeleteDirectoryOutputs = []error{nil}
						Expect(store.DeleteRenditionContent(ctx, userID, imageID, contentID, renditionsID)).To(Succeed())
					})
				})
			})
		})

		Context("Delete", func() {
			AfterEach(func() {
				Expect(underlyingStore.DeleteDirectoryInputs).To(Equal([]string{fmt.Sprintf("%s/%s", userID, imageID)}))
			})

			It("returns an error when the underlying store returns an error", func() {
				parentErr := errorsTest.RandomError()
				underlyingStore.DeleteDirectoryOutputs = []error{parentErr}
				errorsTest.ExpectEqual(store.Delete(ctx, userID, imageID), errors.Wrap(parentErr, "unable to delete image"))
			})

			It("returns successfully when the underlying store returns successfully", func() {
				underlyingStore.DeleteDirectoryOutputs = []error{nil}
				Expect(store.Delete(ctx, userID, imageID)).To(Succeed())
			})
		})

		Context("DeleteAll", func() {
			AfterEach(func() {
				Expect(underlyingStore.DeleteDirectoryInputs).To(Equal([]string{userID}))
			})

			It("returns an error when the underlying store returns an error", func() {
				parentErr := errorsTest.RandomError()
				underlyingStore.DeleteDirectoryOutputs = []error{parentErr}
				errorsTest.ExpectEqual(store.DeleteAll(ctx, userID), errors.Wrap(parentErr, "unable to delete all images"))
			})

			It("returns successfully when the underlying store returns successfully", func() {
				underlyingStore.DeleteDirectoryOutputs = []error{nil}
				Expect(store.DeleteAll(ctx, userID)).To(Succeed())
			})
		})
	})
})
