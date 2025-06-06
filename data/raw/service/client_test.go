package service_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"bytes"
	"context"

	"go.uber.org/mock/gomock"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawService "github.com/tidepool-org/platform/data/raw/service"
	dataRawServiceTest "github.com/tidepool-org/platform/data/raw/service/test"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var (
		controller *gomock.Controller
		mockStore  *dataRawServiceTest.MockStore
	)

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		mockStore = dataRawServiceTest.NewMockStore(controller)
	})

	AfterEach(func() {
		controller.Finish()
	})

	Context("NewClient", func() {
		It("returns error if store is nil", func() {
			clnt, err := dataRawService.NewClient(nil)
			Expect(clnt).To(BeNil())
			Expect(err).To(MatchError("store is missing"))
		})

		It("returns client if store is not nil", func() {
			Expect(dataRawService.NewClient(mockStore)).ToNot(BeNil())
		})
	})

	Context("Client", func() {
		var (
			ctx  context.Context
			clnt *dataRawService.Client
		)

		BeforeEach(func() {
			ctx = context.Background()
			clnt, _ = dataRawService.NewClient(mockStore)
		})

		Context("List", func() {
			It("calls store.List and returns result", func() {
				userID := userTest.RandomUserID()
				filter := &dataRaw.Filter{
					CreatedDate: pointer.FromString(dataRawTest.RandomCreatedDate()),
				}
				pagination := &page.Pagination{Page: test.RandomInt()}
				expected := []*dataRaw.Raw{{ID: test.RandomString()}}
				mockStore.EXPECT().List(ctx, userID, filter, pagination).Return(expected, nil)
				Expect(clnt.List(ctx, userID, filter, pagination)).To(Equal(expected))
			})
		})

		Context("Create", func() {
			It("calls store.Create and returns result", func() {
				userID := userTest.RandomUserID()
				dataSetID := test.RandomString()
				create := &dataRaw.Create{
					DigestMD5: pointer.FromString(test.RandomString()),
				}
				data := bytes.NewBufferString(test.RandomString())
				expected := &dataRaw.Raw{ID: test.RandomString()}
				mockStore.EXPECT().Create(ctx, userID, dataSetID, create, data).Return(expected, nil)
				Expect(clnt.Create(ctx, userID, dataSetID, create, data)).To(Equal(expected))
			})
		})

		Context("Get", func() {
			It("calls store.Get with mapped condition and returns result", func() {
				id := test.RandomString()
				condition := &request.Condition{Revision: pointer.FromInt(test.RandomInt())}
				mappedCondition := storeStructured.MapCondition(condition)
				expected := &dataRaw.Raw{ID: test.RandomString()}
				mockStore.EXPECT().Get(ctx, id, mappedCondition).Return(expected, nil)
				Expect(clnt.Get(ctx, id, condition)).To(Equal(expected))
			})
		})

		Context("GetContent", func() {
			It("calls store.GetContent with mapped condition and returns result", func() {
				id := test.RandomString()
				condition := &request.Condition{Revision: pointer.FromInt(test.RandomInt())}
				mappedCondition := storeStructured.MapCondition(condition)
				expected := &dataRaw.Content{DigestMD5: test.RandomString()}
				mockStore.EXPECT().GetContent(ctx, id, mappedCondition).Return(expected, nil)
				Expect(clnt.GetContent(ctx, id, condition)).To(Equal(expected))
			})
		})

		Context("Update", func() {
			It("calls store.Update with mapped condition and returns result", func() {
				id := test.RandomString()
				condition := &request.Condition{Revision: pointer.FromInt(test.RandomInt())}
				update := &dataRaw.Update{
					ProcessedTime: test.RandomTime(),
				}
				mappedCondition := storeStructured.MapCondition(condition)
				expected := &dataRaw.Raw{ID: test.RandomString()}
				mockStore.EXPECT().Update(ctx, id, mappedCondition, update).Return(expected, nil)
				Expect(clnt.Update(ctx, id, condition, update)).To(Equal(expected))
			})
		})

		Context("Delete", func() {
			It("calls store.Delete with mapped condition and returns result", func() {
				id := test.RandomString()
				revision := test.RandomInt()
				condition := &request.Condition{Revision: &revision}
				mappedCondition := storeStructured.MapCondition(condition)
				expected := &dataRaw.Raw{ID: test.RandomString()}
				mockStore.EXPECT().Delete(ctx, id, mappedCondition).Return(expected, nil)
				Expect(clnt.Delete(ctx, id, condition)).To(Equal(expected))
			})
		})

		Context("DeleteMultiple", func() {
			It("calls store.DeleteMultiple and returns result", func() {
				ids := []string{test.RandomString(), test.RandomString()}
				expected := test.RandomInt()
				mockStore.EXPECT().DeleteMultiple(ctx, ids).Return(expected, nil)
				Expect(clnt.DeleteMultiple(ctx, ids)).To(Equal(expected))
			})
		})

		Context("DeleteAllByDataSetID", func() {
			It("calls store.DeleteAllByDataSetID and returns result", func() {
				userID := userTest.RandomUserID()
				dataSetID := test.RandomString()
				expected := test.RandomInt()
				mockStore.EXPECT().DeleteAllByDataSetID(ctx, userID, dataSetID).Return(expected, nil)
				Expect(clnt.DeleteAllByDataSetID(ctx, userID, dataSetID)).To(Equal(expected))
			})
		})

		Context("DeleteAllByUserID", func() {
			It("calls store.DeleteAllByUserID and returns result", func() {
				userID := userTest.RandomUserID()
				expected := test.RandomInt()
				mockStore.EXPECT().DeleteAllByUserID(ctx, userID).Return(expected, nil)
				Expect(clnt.DeleteAllByUserID(ctx, userID)).To(Equal(expected))
			})
		})
	})
})
