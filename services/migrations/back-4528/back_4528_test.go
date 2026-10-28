package main

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("deleteDeprecatedTasks", func() {
	var ctx context.Context
	var store *storeStructuredMongo.Store
	var repository *storeStructuredMongo.Repository

	deprecatedTypes := []string{
		"org.tidepool.summary.update.cgm",
		"org.tidepool.summary.update.bgm",
		"org.tidepool.summary.migrate.continuous",
	}
	retainedTypes := []string{
		"org.tidepool.ehr.reconcile",
		"org.tidepool.summary",
		"com.example.org.tidepool.summary.update.cgm",
	}

	taskTypes := func() []string {
		cursor, err := repository.Find(ctx, bson.M{})
		Expect(err).ToNot(HaveOccurred())
		var tasks []struct {
			Type string `bson:"type"`
		}
		Expect(cursor.All(ctx, &tasks)).To(Succeed())
		types := []string{}
		for _, task := range tasks {
			types = append(types, task.Type)
		}
		return types
	}

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		store, err = storeStructuredMongo.NewStore(storeStructuredMongoTest.NewConfig())
		Expect(err).ToNot(HaveOccurred())
		repository = store.GetRepository("tasks")

		documents := []any{}
		for _, typ := range append(append([]string{}, deprecatedTypes...), retainedTypes...) {
			documents = append(documents, bson.M{"type": typ})
		}
		_, err = repository.InsertMany(ctx, documents)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if store != nil {
			Expect(store.Terminate(ctx)).To(Succeed())
		}
	})

	It("deletes the tasks of the retired summary runners only", func() {
		count, err := deleteDeprecatedTasks(ctx, repository, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(count).To(Equal(int64(len(deprecatedTypes))))
		Expect(taskTypes()).To(ConsistOf(retainedTypes))
	})

	It("counts the tasks of the retired summary runners without deleting them on a dry run", func() {
		count, err := deleteDeprecatedTasks(ctx, repository, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(count).To(Equal(int64(len(deprecatedTypes))))
		Expect(taskTypes()).To(ConsistOf(append(append([]string{}, deprecatedTypes...), retainedTypes...)))
	})

	It("deletes nothing once the tasks are deleted", func() {
		_, err := deleteDeprecatedTasks(ctx, repository, false)
		Expect(err).ToNot(HaveOccurred())
		count, err := deleteDeprecatedTasks(ctx, repository, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(count).To(BeZero())
	})
})
