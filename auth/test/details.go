package test

import "github.com/tidepool-org/platform/id"

type Details struct {
	ID                  string
	TokenInvocations    int
	TokenOutputs        []string
	IsServerInvocations int
	IsServerOutputs     []bool
	UserIDInvocations   int
	UserIDOutputs       []string
}

func NewDetails() *Details {
	return &Details{
		ID: id.New(),
	}
}

func (d *Details) Token() string {
	d.TokenInvocations++

	if len(d.TokenOutputs) == 0 {
		panic("Unexpected invocation of Token on Details")
	}

	output := d.TokenOutputs[0]
	d.TokenOutputs = d.TokenOutputs[1:]
	return output
}

func (d *Details) IsServer() bool {
	d.IsServerInvocations++

	if len(d.IsServerOutputs) == 0 {
		panic("Unexpected invocation of IsServer on Details")
	}

	output := d.IsServerOutputs[0]
	d.IsServerOutputs = d.IsServerOutputs[1:]
	return output
}

func (d *Details) UserID() string {
	d.UserIDInvocations++

	if len(d.UserIDOutputs) == 0 {
		panic("Unexpected invocation of UserID on Details")
	}

	output := d.UserIDOutputs[0]
	d.UserIDOutputs = d.UserIDOutputs[1:]
	return output
}

func (d *Details) UnusedOutputsCount() int {
	return len(d.TokenOutputs) +
		len(d.IsServerOutputs) +
		len(d.UserIDOutputs)
}
