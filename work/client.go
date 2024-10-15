package work

import (
	"context"
)

//go:generate mockgen --build_flags=--mod=mod -source=./client.go -destination=./test/client.go -package test Client
type Client interface {
	Create(ctx context.Context, create *Create) (*Work, error)
	Get(ctx context.Context, id string) (*Work, error)
	Process(ctx context.Context, process *Process) (*Work, error)
	Repeat(ctx context.Context, id string, update *Repeat) (*Work, error)
	Delete(ctx context.Context, id string) error
}
