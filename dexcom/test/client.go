package test

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/oauth"
)

type GetAlertsInput struct {
	StartTime   time.Time
	EndTime     time.Time
	TokenSource oauth.TokenSource
}

type GetAlertsOutput struct {
	AlertsResponse *dexcom.AlertsResponse
	Error          error
}

type GetCalibrationsInput struct {
	StartTime   time.Time
	EndTime     time.Time
	TokenSource oauth.TokenSource
}

type GetCalibrationsOutput struct {
	CalibrationsResponse *dexcom.CalibrationsResponse
	Error                error
}

type GetDataRangeInput struct {
	LastSyncTime *time.Time
	TokenSource  oauth.TokenSource
}

type GetDataRangeOutput struct {
	DataRangeResponse *dexcom.DataRangesResponse
	Error             error
}

type GetDevicesInput struct {
	StartTime   time.Time
	EndTime     time.Time
	TokenSource oauth.TokenSource
}

type GetDevicesOutput struct {
	DevicesResponse *dexcom.DevicesResponse
	Error           error
}

type GetEGVsInput struct {
	StartTime   time.Time
	EndTime     time.Time
	TokenSource oauth.TokenSource
}

type GetEGVsOutput struct {
	EGVsResponse *dexcom.EGVsResponse
	Error        error
}

type GetEventsInput struct {
	StartTime   time.Time
	EndTime     time.Time
	TokenSource oauth.TokenSource
}

type GetEventsOutput struct {
	EventsResponse *dexcom.EventsResponse
	Error          error
}

type Client struct {
	GetAlertsInvocations       int
	GetAlertsInputs            []GetAlertsInput
	GetAlertsStub              func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.AlertsResponse, error)
	GetAlertsOutputs           []GetAlertsOutput
	GetAlertsOutput            *GetAlertsOutput
	GetCalibrationsInvocations int
	GetCalibrationsInputs      []GetCalibrationsInput
	GetCalibrationsStub        func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.CalibrationsResponse, error)
	GetCalibrationsOutputs     []GetCalibrationsOutput
	GetCalibrationsOutput      *GetCalibrationsOutput
	GetDataRangeInvocations    int
	GetDataRangeInputs         []GetDataRangeInput
	GetDataRangeStub           func(ctx context.Context, lastSyncTime *time.Time, tokenSource oauth.TokenSource) (*dexcom.DataRangesResponse, error)
	GetDataRangeOutputs        []GetDataRangeOutput
	GetDataRangeOutput         *GetDataRangeOutput
	GetDevicesInvocations      int
	GetDevicesInputs           []GetDevicesInput
	GetDevicesStub             func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.DevicesResponse, error)
	GetDevicesOutputs          []GetDevicesOutput
	GetDevicesOutput           *GetDevicesOutput
	GetEGVsInvocations         int
	GetEGVsInputs              []GetEGVsInput
	GetEGVsStub                func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EGVsResponse, error)
	GetEGVsOutputs             []GetEGVsOutput
	GetEGVsOutput              *GetEGVsOutput
	GetEventsInvocations       int
	GetEventsInputs            []GetEventsInput
	GetEventsStub              func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EventsResponse, error)
	GetEventsOutputs           []GetEventsOutput
	GetEventsOutput            *GetEventsOutput
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) GetAlerts(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.AlertsResponse, error) {
	c.GetAlertsInvocations++
	c.GetAlertsInputs = append(c.GetAlertsInputs, GetAlertsInput{StartTime: startTime, EndTime: endTime, TokenSource: tokenSource})
	if c.GetAlertsStub != nil {
		return c.GetAlertsStub(ctx, startTime, endTime, tokenSource)
	}
	if len(c.GetAlertsOutputs) > 0 {
		output := c.GetAlertsOutputs[0]
		c.GetAlertsOutputs = c.GetAlertsOutputs[1:]
		return output.AlertsResponse, output.Error
	}
	if c.GetAlertsOutput != nil {
		return c.GetAlertsOutput.AlertsResponse, c.GetAlertsOutput.Error
	}
	panic("GetAlerts has no output")
}

func (c *Client) GetCalibrations(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.CalibrationsResponse, error) {
	c.GetCalibrationsInvocations++
	c.GetCalibrationsInputs = append(c.GetCalibrationsInputs, GetCalibrationsInput{StartTime: startTime, EndTime: endTime, TokenSource: tokenSource})
	if c.GetCalibrationsStub != nil {
		return c.GetCalibrationsStub(ctx, startTime, endTime, tokenSource)
	}
	if len(c.GetCalibrationsOutputs) > 0 {
		output := c.GetCalibrationsOutputs[0]
		c.GetCalibrationsOutputs = c.GetCalibrationsOutputs[1:]
		return output.CalibrationsResponse, output.Error
	}
	if c.GetCalibrationsOutput != nil {
		return c.GetCalibrationsOutput.CalibrationsResponse, c.GetCalibrationsOutput.Error
	}
	panic("GetCalibrations has no output")
}

func (c *Client) GetDataRange(ctx context.Context, lastSyncTime *time.Time, tokenSource oauth.TokenSource) (*dexcom.DataRangesResponse, error) {
	c.GetDataRangeInvocations++
	c.GetDataRangeInputs = append(c.GetDataRangeInputs, GetDataRangeInput{LastSyncTime: lastSyncTime, TokenSource: tokenSource})
	if c.GetDataRangeStub != nil {
		return c.GetDataRangeStub(ctx, lastSyncTime, tokenSource)
	}
	if len(c.GetDataRangeOutputs) > 0 {
		output := c.GetDataRangeOutputs[0]
		c.GetDataRangeOutputs = c.GetDataRangeOutputs[1:]
		return output.DataRangeResponse, output.Error
	}
	if c.GetDataRangeOutput != nil {
		return c.GetDataRangeOutput.DataRangeResponse, c.GetDataRangeOutput.Error
	}
	panic("GetDataRange has no output")
}

func (c *Client) GetDevices(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.DevicesResponse, error) {
	c.GetDevicesInvocations++
	c.GetDevicesInputs = append(c.GetDevicesInputs, GetDevicesInput{StartTime: startTime, EndTime: endTime, TokenSource: tokenSource})
	if c.GetDevicesStub != nil {
		return c.GetDevicesStub(ctx, startTime, endTime, tokenSource)
	}
	if len(c.GetDevicesOutputs) > 0 {
		output := c.GetDevicesOutputs[0]
		c.GetDevicesOutputs = c.GetDevicesOutputs[1:]
		return output.DevicesResponse, output.Error
	}
	if c.GetDevicesOutput != nil {
		return c.GetDevicesOutput.DevicesResponse, c.GetDevicesOutput.Error
	}
	panic("GetDevices has no output")
}

func (c *Client) GetEGVs(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EGVsResponse, error) {
	c.GetEGVsInvocations++
	c.GetEGVsInputs = append(c.GetEGVsInputs, GetEGVsInput{StartTime: startTime, EndTime: endTime, TokenSource: tokenSource})
	if c.GetEGVsStub != nil {
		return c.GetEGVsStub(ctx, startTime, endTime, tokenSource)
	}
	if len(c.GetEGVsOutputs) > 0 {
		output := c.GetEGVsOutputs[0]
		c.GetEGVsOutputs = c.GetEGVsOutputs[1:]
		return output.EGVsResponse, output.Error
	}
	if c.GetEGVsOutput != nil {
		return c.GetEGVsOutput.EGVsResponse, c.GetEGVsOutput.Error
	}
	panic("GetEGVs has no output")
}

func (c *Client) GetEvents(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EventsResponse, error) {
	c.GetEventsInvocations++
	c.GetEventsInputs = append(c.GetEventsInputs, GetEventsInput{StartTime: startTime, EndTime: endTime, TokenSource: tokenSource})
	if c.GetEventsStub != nil {
		return c.GetEventsStub(ctx, startTime, endTime, tokenSource)
	}
	if len(c.GetEventsOutputs) > 0 {
		output := c.GetEventsOutputs[0]
		c.GetEventsOutputs = c.GetEventsOutputs[1:]
		return output.EventsResponse, output.Error
	}
	if c.GetEventsOutput != nil {
		return c.GetEventsOutput.EventsResponse, c.GetEventsOutput.Error
	}
	panic("GetEvents has no output")
}

func (c *Client) AssertOutputsEmpty() {
	if len(c.GetAlertsOutputs) > 0 {
		panic("GetAlertsOutputs is not empty")
	}
	if len(c.GetCalibrationsOutputs) > 0 {
		panic("GetCalibrationsOutputs is not empty")
	}
	if len(c.GetDataRangeOutputs) > 0 {
		panic("GetDataRangeOutputs is not empty")
	}
	if len(c.GetDevicesOutputs) > 0 {
		panic("GetDevicesOutputs is not empty")
	}
	if len(c.GetEGVsOutputs) > 0 {
		panic("GetEGVsOutputs is not empty")
	}
	if len(c.GetEventsOutputs) > 0 {
		panic("GetEventsOutputs is not empty")
	}
}
