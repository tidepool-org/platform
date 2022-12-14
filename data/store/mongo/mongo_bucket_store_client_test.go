package mongo_test

import (
	"reflect"
	"testing"
	"time"

	goComMgo "github.com/mdblp/go-common/clients/mongo"
	log "github.com/sirupsen/logrus"

	"github.com/tidepool-org/platform/data/schema"
	"github.com/tidepool-org/platform/data/store/mongo"
)

func TestMongoBucketStoreClient_BuildUserMetadata(t *testing.T) {
	type args struct {
		incomingUserMetadata *schema.Metadata
		creationTimestamp    time.Time
		strUserId            string
		dataTimestamp        time.Time
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
				incomingUserMetadata: nil,
				creationTimestamp:    testTime,
				strUserId:            "123456789",
				dataTimestamp:        testTime,
			},
			want: &schema.Metadata{
				Id:                  "",
				CreationTimestamp:   testTime,
				UserId:              "123456789",
				OldestDataTimestamp: testTime,
				NewestDataTimestamp: testTime,
			},
		},
		{
			name: "given a 70's data timestamp should not update oldest metadata",
			args: args{
				incomingUserMetadata: &schema.Metadata{
					Id:                  "metadata1234",
					CreationTimestamp:   testTime,
					UserId:              "123456789",
					OldestDataTimestamp: testTime,
					NewestDataTimestamp: testTime,
				},
				creationTimestamp: testTime,
				strUserId:         "123456789",
				dataTimestamp:     veryOldTime,
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
				incomingUserMetadata: &schema.Metadata{
					Id:                  "metadata1234",
					CreationTimestamp:   testTime,
					UserId:              "123456789",
					OldestDataTimestamp: testTime,
					NewestDataTimestamp: testTime,
				},
				creationTimestamp: testTime,
				strUserId:         "123456789",
				dataTimestamp:     beforeTestTime,
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
				incomingUserMetadata: &schema.Metadata{
					Id:                  "metadata1234",
					CreationTimestamp:   beforeTestTime,
					UserId:              "123456789",
					OldestDataTimestamp: beforeTestTime,
					NewestDataTimestamp: beforeTestTime,
				},
				creationTimestamp: testTime,
				strUserId:         "123456789",
				dataTimestamp:     testTime,
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
			if got := c.BuildUserMetadata(tt.args.incomingUserMetadata, tt.args.creationTimestamp, tt.args.strUserId, tt.args.dataTimestamp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildUserMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
