package metric

import "context"

type Client interface {
	RecordMetric(ctx context.Context, name string, data ...map[string]string) error
}
