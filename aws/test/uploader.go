package test

import (
	awsSdkGoAws "github.com/aws/aws-sdk-go/aws"
	awsSdkGoServiceS3S3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
	awsSdkGoServiceS3S3managerS3manageriface "github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type UploadWithContextInput struct {
	Input   *awsSdkGoServiceS3S3manager.UploadInput
	Options []func(*awsSdkGoServiceS3S3manager.Uploader)
}

type UploadWithContextOutput struct {
	Output *awsSdkGoServiceS3S3manager.UploadOutput
	Error  error
}

type Uploader struct {
	awsSdkGoServiceS3S3managerS3manageriface.UploaderAPI

	UploadWithContextInvocations int
	UploadWithContextInputs      []UploadWithContextInput
	UploadWithContextStub        func(ctx awsSdkGoAws.Context, input *awsSdkGoServiceS3S3manager.UploadInput, options ...func(*awsSdkGoServiceS3S3manager.Uploader)) (*awsSdkGoServiceS3S3manager.UploadOutput, error)
	UploadWithContextOutputs     []UploadWithContextOutput
	UploadWithContextOutput      *UploadWithContextOutput
}

func NewUploader() *Uploader {
	return &Uploader{}
}

func (u *Uploader) UploadWithContext(ctx awsSdkGoAws.Context, input *awsSdkGoServiceS3S3manager.UploadInput, options ...func(*awsSdkGoServiceS3S3manager.Uploader)) (*awsSdkGoServiceS3S3manager.UploadOutput, error) {
	u.UploadWithContextInvocations++
	u.UploadWithContextInputs = append(u.UploadWithContextInputs, UploadWithContextInput{Input: input, Options: options})
	if u.UploadWithContextStub != nil {
		return u.UploadWithContextStub(ctx, input, options...)
	}
	if len(u.UploadWithContextOutputs) > 0 {
		output := u.UploadWithContextOutputs[0]
		u.UploadWithContextOutputs = u.UploadWithContextOutputs[1:]
		return output.Output, output.Error
	}
	if u.UploadWithContextOutput != nil {
		return u.UploadWithContextOutput.Output, u.UploadWithContextOutput.Error
	}
	panic("UploadWithContext has no output")
}

func (u *Uploader) SetUploadWithContextOutput(output UploadWithContextOutput) {
	u.UploadWithContextOutput = &output
}

func (u *Uploader) AssertOutputsEmpty() {
	if len(u.UploadWithContextOutputs) > 0 {
		panic("UploadWithContextOutputs is not empty")
	}
}
