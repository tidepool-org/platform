package unstructured_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	blobStoreUnstructured "github.com/tidepool-org/platform/blob/store/unstructured"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	storeUnstructuredTest "github.com/tidepool-org/platform/store/unstructured/test"
	"github.com/tidepool-org/platform/test"
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
			store, err := blobStoreUnstructured.NewStore(nil)
			errorsTest.ExpectEqual(err, errors.New("store is missing"))
			Expect(store).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(blobStoreUnstructured.NewStore(underlyingStore)).ToNot(BeNil())
		})
	})

	Context("with new store", func() {
		var store *blobStoreUnstructured.StoreImpl
		var ctx context.Context
		var userID string
		var id string
		var key string

		BeforeEach(func() {
			var err error
			store, err = blobStoreUnstructured.NewStore(underlyingStore)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			ctx = context.Background()
			userID = test.RandomString()
			id = test.RandomString()
			key = fmt.Sprintf("%s/%s/%s", userID, id, id)
		})

		Context("Exists", func() {
			AfterEach(func() {
				Expect(underlyingStore.ExistsInputs).To(Equal([]storeUnstructuredTest.ExistsInput{{Context: ctx, Key: key}}))
			})

			It("returns an error when the underlying store returns an error", func() {
				parentErr := errorsTest.NewError()
				underlyingStore.ExistsOutputs = []storeUnstructuredTest.ExistsOutput{{Exists: false, Error: parentErr}}
				exists, err := store.Exists(ctx, userID, id)
				errorsTest.ExpectEqual(err, errors.New("unable to exists blob"))
				Expect(exists).To(BeFalse())
			})

			It("returns false when the underlying store returns false", func() {
				underlyingStore.ExistsOutputs = []storeUnstructuredTest.ExistsOutput{{Exists: false, Error: nil}}
				Expect(store.Exists(ctx, userID, id)).To(BeFalse())
			})

			It("returns true when the underlying store returns true", func() {
				underlyingStore.ExistsOutputs = []storeUnstructuredTest.ExistsOutput{{Exists: true, Error: nil}}
				Expect(store.Exists(ctx, userID, id)).To(BeTrue())
			})
		})

		Context("Put", func() {
			var reader io.Reader

			BeforeEach(func() {
				reader = strings.NewReader(test.RandomString())
			})

			AfterEach(func() {
				Expect(underlyingStore.PutInputs).To(Equal([]storeUnstructuredTest.PutInput{{Context: ctx, Key: key, Reader: reader}}))
			})

			It("returns an error when the underlying store returns an error", func() {
				parentErr := errorsTest.NewError()
				underlyingStore.PutOutputs = []error{parentErr}
				errorsTest.ExpectEqual(store.Put(ctx, userID, id, reader), errors.New("unable to put blob"))
			})

			It("returns successfully when the underlying store returns successfully", func() {
				underlyingStore.PutOutputs = []error{nil}
				Expect(store.Put(ctx, userID, id, reader)).ToNot(HaveOccurred())
			})
		})

		Context("Get", func() {
			var parentReader io.ReadCloser

			BeforeEach(func() {
				parentReader = ioutil.NopCloser(strings.NewReader(test.RandomString()))
			})

			AfterEach(func() {
				Expect(underlyingStore.GetInputs).To(Equal([]storeUnstructuredTest.GetInput{{Context: ctx, Key: key}}))
			})

			It("returns an error when the underlying store returns an error", func() {
				parentErr := errorsTest.NewError()
				underlyingStore.GetOutputs = []storeUnstructuredTest.GetOutput{{Reader: nil, Error: parentErr}}
				reader, err := store.Get(ctx, userID, id)
				errorsTest.ExpectEqual(err, errors.New("unable to get blob"))
				Expect(reader).To(BeNil())
			})

			It("returns nil when the underlying store returns nil", func() {
				underlyingStore.GetOutputs = []storeUnstructuredTest.GetOutput{{Reader: nil, Error: nil}}
				Expect(store.Get(ctx, userID, id)).To(BeNil())
			})

			It("returns a reader when the underlying store returns a reader", func() {
				underlyingStore.GetOutputs = []storeUnstructuredTest.GetOutput{{Reader: parentReader, Error: nil}}
				Expect(store.Get(ctx, userID, id)).To(Equal(parentReader))
			})
		})

		Context("Delete", func() {
			AfterEach(func() {
				Expect(underlyingStore.DeleteInputs).To(Equal([]storeUnstructuredTest.DeleteInput{{Context: ctx, Key: key}}))
			})

			It("returns an error when the underlying store returns an error", func() {
				parentErr := errorsTest.NewError()
				underlyingStore.DeleteOutputs = []storeUnstructuredTest.DeleteOutput{{Deleted: false, Error: parentErr}}
				exists, err := store.Delete(ctx, userID, id)
				errorsTest.ExpectEqual(err, errors.New("unable to delete blob"))
				Expect(exists).To(BeFalse())
			})

			It("returns false when the underlying store returns false", func() {
				underlyingStore.DeleteOutputs = []storeUnstructuredTest.DeleteOutput{{Deleted: false, Error: nil}}
				Expect(store.Delete(ctx, userID, id)).To(BeFalse())
			})

			It("returns true when the underlying store returns true", func() {
				underlyingStore.DeleteOutputs = []storeUnstructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
				Expect(store.Delete(ctx, userID, id)).To(BeTrue())
			})
		})
	})
})
