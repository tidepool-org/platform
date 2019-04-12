package test

import (
	awsSdkGoServiceS3S3iface "github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/tidepool-org/platform/aws"
)

type API struct {
	S3Invocations        int
	S3Stub               func() awsSdkGoServiceS3S3iface.S3API
	S3Outputs            []awsSdkGoServiceS3S3iface.S3API
	S3Output             *awsSdkGoServiceS3S3iface.S3API
	S3ManagerInvocations int
	S3ManagerStub        func() aws.S3Manager
	S3ManagerOutputs     []aws.S3Manager
	S3ManagerOutput      *aws.S3Manager
}

func NewAPI() *API {
	return &API{}
}

func (a *API) S3() awsSdkGoServiceS3S3iface.S3API {
	a.S3Invocations++
	if a.S3Stub != nil {
		return a.S3Stub()
	}
	if len(a.S3Outputs) > 0 {
		output := a.S3Outputs[0]
		a.S3Outputs = a.S3Outputs[1:]
		return output
	}
	if a.S3Output != nil {
		return *a.S3Output
	}
	panic("S3 has no output")
}

func (a *API) S3Manager() aws.S3Manager {
	a.S3ManagerInvocations++
	if a.S3ManagerStub != nil {
		return a.S3ManagerStub()
	}
	if len(a.S3ManagerOutputs) > 0 {
		output := a.S3ManagerOutputs[0]
		a.S3ManagerOutputs = a.S3ManagerOutputs[1:]
		return output
	}
	if a.S3ManagerOutput != nil {
		return *a.S3ManagerOutput
	}
	panic("S3ManagerAPI has no output")
}

func (a *API) SetS3Output(output awsSdkGoServiceS3S3iface.S3API) {
	a.S3Output = &output
}

func (a *API) SetS3ManagerOutput(output aws.S3Manager) {
	a.S3ManagerOutput = &output
}

func (a *API) AssertOutputsEmpty() {
	if len(a.S3Outputs) > 0 {
		panic("S3Outputs is not empty")
	}
	if len(a.S3ManagerOutputs) > 0 {
		panic("S3ManagerOutputs is not empty")
	}
}
