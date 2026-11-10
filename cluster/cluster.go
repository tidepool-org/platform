package cluster

import (
	"context"
	"time"
)

//go:generate mockgen -source=cluster.go -destination=test/cluster_mocks.go -package=test -typed

// Release is the cluster configuration repository commit in service and when its configuration was first applied
type Release struct {
	Revision string
	Time     time.Time
}

type ReleaseReporter interface {
	Release(ctx context.Context) (*Release, error)
}
