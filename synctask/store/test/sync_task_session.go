package test

import (
	"context"

	"github.com/onsi/gomega"
	testStore "github.com/tidepool-org/platform/store/test"
)

type DestroySyncTasksForUserByIDInput struct {
	Context context.Context
	UserID  string
}

type SyncTaskSession struct {
	*testStore.Session
	DestroySyncTasksForUserByIDInvocations int
	DestroySyncTasksForUserByIDInputs      []DestroySyncTasksForUserByIDInput
	DestroySyncTasksForUserByIDOutputs     []error
}

func NewSyncTaskSession() *SyncTaskSession {
	return &SyncTaskSession{
		Session: testStore.NewSession(),
	}
}

func (s *SyncTaskSession) DestroySyncTasksForUserByID(ctx context.Context, userID string) error {
	s.DestroySyncTasksForUserByIDInvocations++

	s.DestroySyncTasksForUserByIDInputs = append(s.DestroySyncTasksForUserByIDInputs, DestroySyncTasksForUserByIDInput{Context: ctx, UserID: userID})

	gomega.Expect(s.DestroySyncTasksForUserByIDOutputs).ToNot(gomega.BeEmpty())

	output := s.DestroySyncTasksForUserByIDOutputs[0]
	s.DestroySyncTasksForUserByIDOutputs = s.DestroySyncTasksForUserByIDOutputs[1:]
	return output
}

func (s *SyncTaskSession) Expectations() {
	s.Session.Expectations()
	gomega.Expect(s.DestroySyncTasksForUserByIDOutputs).To(gomega.BeEmpty())
}
