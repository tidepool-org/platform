package test

import (
	"context"

	"github.com/onsi/gomega"
	"github.com/tidepool-org/platform/test"
)

type DestroySyncTasksForUserByIDInput struct {
	Context context.Context
	UserID  string
}

type SyncTaskSession struct {
	*test.Closer
	DestroySyncTasksForUserByIDInvocations int
	DestroySyncTasksForUserByIDInputs      []DestroySyncTasksForUserByIDInput
	DestroySyncTasksForUserByIDOutputs     []error
}

func NewSyncTaskSession() *SyncTaskSession {
	return &SyncTaskSession{
		Closer: test.NewCloser(),
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
	s.Closer.AssertOutputsEmpty()
	gomega.Expect(s.DestroySyncTasksForUserByIDOutputs).To(gomega.BeEmpty())
}
