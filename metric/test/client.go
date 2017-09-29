package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

type RecordMetricInput struct {
	Context context.Context
	Name    string
	Data    []map[string]string
}

type Client struct {
	*test.Mock
	RecordMetricInvocations int
	RecordMetricInputs      []RecordMetricInput
	RecordMetricOutputs     []error
}

func NewClient() *Client {
	return &Client{
		Mock: test.NewMock(),
	}
}

func (c *Client) RecordMetric(ctx context.Context, name string, data ...map[string]string) error {
	c.RecordMetricInvocations++

	c.RecordMetricInputs = append(c.RecordMetricInputs, RecordMetricInput{Context: ctx, Name: name, Data: data})

	gomega.Expect(c.RecordMetricOutputs).ToNot(gomega.BeEmpty())

	output := c.RecordMetricOutputs[0]
	c.RecordMetricOutputs = c.RecordMetricOutputs[1:]
	return output
}

func (c *Client) Expectations() {
	c.Mock.Expectations()
	gomega.Expect(c.RecordMetricOutputs).To(gomega.BeEmpty())
}
