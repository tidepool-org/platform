package test

import (
	"context"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

type ListDeviceLogsInput struct {
	UserID     string
	Filter     *blob.DeviceLogsFilter
	Pagination *page.Pagination
}

type ListDeviceLogsOutput struct {
	DeviceLogs blob.DeviceLogsBlobArray
	Error      error
}

type CreateDeviceLogsInput struct {
	UserID string
	Create *blobStoreStructured.Create
}

type CreateDeviceLogsOutput struct {
	DeviceLogsBlob *blob.DeviceLogsBlob
	Error          error
}

type UpdateDeviceLogsInput struct {
	ID        string
	Condition *request.Condition
	Update    *blobStoreStructured.DeviceLogsUpdate
}

type UpdateDeviceLogsOutput struct {
	DeviceLogsBlob *blob.DeviceLogsBlob
	Error          error
}

type DestroyDeviceLogsInput struct {
	ID        string
	Condition *request.Condition
}

type DestroyDeviceLogsOutput struct {
	Destroyed bool
	Error     error
}

type DeviceLogsRepository struct {
	ListInvocations int
	ListInputs      []ListDeviceLogsInput
	ListStub        func(ctx context.Context, userID string, filter *blob.DeviceLogsFilter, pagination *page.Pagination) (blob.DeviceLogsBlobArray, error)
	ListOutputs     []ListDeviceLogsOutput
	ListOutput      *ListDeviceLogsOutput

	CreateInvocations int
	CreateInputs      []CreateDeviceLogsInput
	CreateStub        func(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.DeviceLogsBlob, error)
	CreateOutputs     []CreateDeviceLogsOutput
	CreateOutput      *CreateDeviceLogsOutput

	UpdateInvocations int
	UpdateInputs      []UpdateDeviceLogsInput
	UpdateStub        func(ctx context.Context, id string, condition *request.Condition, update *blobStoreStructured.DeviceLogsUpdate) (*blob.DeviceLogsBlob, error)
	UpdateOutputs     []UpdateDeviceLogsOutput
	UpdateOutput      *UpdateDeviceLogsOutput

	DestroyInvocations int
	DestroyInputs      []DestroyDeviceLogsInput
	DestroyStub        func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DestroyOutputs     []DestroyDeviceLogsOutput
	DestroyOutput      *DestroyDeviceLogsOutput
}

func NewDeviceLogsRepository() *DeviceLogsRepository {
	return &DeviceLogsRepository{}
}

func (d *DeviceLogsRepository) List(ctx context.Context, userID string, filter *blob.DeviceLogsFilter, pagination *page.Pagination) (blob.DeviceLogsBlobArray, error) {
	d.ListInvocations++
	d.ListInputs = append(d.ListInputs, ListDeviceLogsInput{UserID: userID, Filter: filter, Pagination: pagination})
	if d.ListStub != nil {
		return d.ListStub(ctx, userID, filter, pagination)
	}
	if len(d.ListOutputs) > 0 {
		output := d.ListOutputs[0]
		d.ListOutputs = d.ListOutputs[1:]
		return output.DeviceLogs, output.Error
	}
	if d.ListOutput != nil {
		return d.ListOutput.DeviceLogs, d.ListOutput.Error
	}
	panic("List has no output")
}

func (d *DeviceLogsRepository) Get(ctx context.Context, deviceLogID string) (*blob.DeviceLogsBlob, error) {
	return nil, nil
}

func (d *DeviceLogsRepository) Create(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.DeviceLogsBlob, error) {
	d.CreateInvocations++
	d.CreateInputs = append(d.CreateInputs, CreateDeviceLogsInput{UserID: userID, Create: create})
	if d.CreateStub != nil {
		return d.CreateStub(ctx, userID, create)
	}
	if len(d.CreateOutputs) > 0 {
		output := d.CreateOutputs[0]
		d.CreateOutputs = d.CreateOutputs[1:]
		return output.DeviceLogsBlob, output.Error
	}
	if d.CreateOutput != nil {
		return d.CreateOutput.DeviceLogsBlob, d.CreateOutput.Error
	}
	panic("Create has no output")
}

func (d *DeviceLogsRepository) Update(ctx context.Context, id string, condition *request.Condition, update *blobStoreStructured.DeviceLogsUpdate) (*blob.DeviceLogsBlob, error) {
	d.UpdateInvocations++
	d.UpdateInputs = append(d.UpdateInputs, UpdateDeviceLogsInput{ID: id, Condition: condition, Update: update})
	if d.UpdateStub != nil {
		return d.UpdateStub(ctx, id, condition, update)
	}
	if len(d.UpdateOutputs) > 0 {
		output := d.UpdateOutputs[0]
		d.UpdateOutputs = d.UpdateOutputs[1:]
		return output.DeviceLogsBlob, output.Error
	}
	if d.UpdateOutput != nil {
		return d.UpdateOutput.DeviceLogsBlob, d.UpdateOutput.Error
	}
	panic("Update has no output")
}

func (d *DeviceLogsRepository) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	d.DestroyInvocations++
	d.DestroyInputs = append(d.DestroyInputs, DestroyDeviceLogsInput{ID: id, Condition: condition})
	if d.DestroyStub != nil {
		return d.DestroyStub(ctx, id, condition)
	}
	if len(d.DestroyOutputs) > 0 {
		output := d.DestroyOutputs[0]
		d.DestroyOutputs = d.DestroyOutputs[1:]
		return output.Destroyed, output.Error
	}
	if d.DestroyOutput != nil {
		return d.DestroyOutput.Destroyed, d.DestroyOutput.Error
	}
	panic("Destroy has no output")
}

func (b *DeviceLogsRepository) AssertOutputsEmpty() {
	if len(b.CreateOutputs) > 0 {
		panic("CreateOutputs is not empty")
	}
	if len(b.UpdateOutputs) > 0 {
		panic("UpdateOutputs is not empty")
	}
	if len(b.DestroyOutputs) > 0 {
		panic("DestroyOutputs is not empty")
	}
}
