package mongo_test

import (
	"reflect"
	"testing"
	"time"

	goComMgo "github.com/mdblp/go-db/mongo"
	log "github.com/sirupsen/logrus"

	"github.com/tidepool-org/platform/data/schema"
	"github.com/tidepool-org/platform/data/store/mongo"
)

func TestMongoBucketStoreClient_BuildUserMetadata(t *testing.T) {
	type args struct {
		incomingUserMetadata *schema.Metadata
		dbUserMetadata       *schema.Metadata
	}
	testTime := time.Now()
	beforeTestTime := testTime.Add(-24 * time.Hour)
	veryOldTime := testTime.Add(-30 * 365 * 24 * time.Hour)
	tests := []struct {
		name string
		args args
		want *schema.Metadata
	}{
		{
			name: "given empty user metadata should create a new one with passed params",
			args: args{
				dbUserMetadata: nil,
				incomingUserMetadata: &schema.Metadata{
					Id:                  "test1234",
					CreationTimestamp:   testTime,
					UserId:              "123456789",
					OldestDataTimestamp: testTime,
					NewestDataTimestamp: testTime,
				},
			},
			want: &schema.Metadata{
				Id:                  "test1234",
				CreationTimestamp:   testTime,
				UserId:              "123456789",
				OldestDataTimestamp: testTime,
				NewestDataTimestamp: testTime,
			},
		},
		{
			name: "given a 70's data timestamp should not update oldest metadata",
			args: args{
				dbUserMetadata: &schema.Metadata{
					Id:                  "metadata1234",
					CreationTimestamp:   testTime,
					UserId:              "123456789",
					OldestDataTimestamp: testTime,
					NewestDataTimestamp: testTime,
				},
				incomingUserMetadata: &schema.Metadata{
					Id:                  "metadata1234",
					CreationTimestamp:   veryOldTime,
					UserId:              "123456789",
					OldestDataTimestamp: veryOldTime,
					NewestDataTimestamp: veryOldTime,
				},
			},
			want: &schema.Metadata{
				Id:                  "metadata1234",
				CreationTimestamp:   testTime,
				UserId:              "123456789",
				OldestDataTimestamp: testTime,
				NewestDataTimestamp: testTime,
			},
		},
		{
			name: "given a normal data timestamp should update oldest metadata",
			args: args{
				dbUserMetadata: &schema.Metadata{
					Id:                  "metadata1234",
					CreationTimestamp:   testTime,
					UserId:              "123456789",
					OldestDataTimestamp: testTime,
					NewestDataTimestamp: testTime,
				},
				incomingUserMetadata: &schema.Metadata{
					Id:                  "metadata1234",
					CreationTimestamp:   beforeTestTime,
					UserId:              "123456789",
					OldestDataTimestamp: beforeTestTime,
					NewestDataTimestamp: beforeTestTime,
				},
			},
			want: &schema.Metadata{
				Id:                  "metadata1234",
				CreationTimestamp:   testTime,
				UserId:              "123456789",
				OldestDataTimestamp: beforeTestTime,
				NewestDataTimestamp: testTime,
			},
		},
		{
			name: "given a normal data timestamp should update newest metadata",
			args: args{
				dbUserMetadata: &schema.Metadata{
					Id:                  "metadata1234",
					CreationTimestamp:   beforeTestTime,
					UserId:              "123456789",
					OldestDataTimestamp: beforeTestTime,
					NewestDataTimestamp: beforeTestTime,
				},
				incomingUserMetadata: &schema.Metadata{
					Id:                  "metadata1234",
					CreationTimestamp:   testTime,
					UserId:              "123456789",
					OldestDataTimestamp: testTime,
					NewestDataTimestamp: testTime,
				},
			},
			want: &schema.Metadata{
				Id:                  "metadata1234",
				CreationTimestamp:   beforeTestTime,
				UserId:              "123456789",
				OldestDataTimestamp: beforeTestTime,
				NewestDataTimestamp: testTime,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := mongo.NewMongoBucketStoreClient(&goComMgo.Config{}, &log.Logger{}, 2015)
			if got, _ := c.RefreshUserMetadata(tt.args.dbUserMetadata, tt.args.incomingUserMetadata); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildUserMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
