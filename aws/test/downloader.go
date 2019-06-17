package test

import (
	"io"

	awsSdkGoAws "github.com/aws/aws-sdk-go/aws"
	awsSdkGoServiceS3 "github.com/aws/aws-sdk-go/service/s3"
	awsSdkGoServiceS3S3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
	awsSdkGoServiceS3S3managerS3manageriface "github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type DownloadWithContextInput struct {
	WriterAt io.WriterAt
	Input    *awsSdkGoServiceS3.GetObjectInput
	Options  []func(*awsSdkGoServiceS3S3manager.Downloader)
}

type DownloadWithContextOutput struct {
	BytesWritten int64
	Error        error
}

type Downloader struct {
	awsSdkGoServiceS3S3managerS3manageriface.DownloaderAPI

	DownloadWithContextInvocations int
	DownloadWithContextInputs      []DownloadWithContextInput
	DownloadWithContextStub        func(ctx awsSdkGoAws.Context, writerAt io.WriterAt, input *awsSdkGoServiceS3.GetObjectInput, options ...func(*awsSdkGoServiceS3S3manager.Downloader)) (int64, error)
	DownloadWithContextOutputs     []DownloadWithContextOutput
	DownloadWithContextOutput      *DownloadWithContextOutput
}

func NewDownloader() *Downloader {
	return &Downloader{}
}

func (d *Downloader) DownloadWithContext(ctx awsSdkGoAws.Context, writerAt io.WriterAt, input *awsSdkGoServiceS3.GetObjectInput, options ...func(*awsSdkGoServiceS3S3manager.Downloader)) (int64, error) {
	d.DownloadWithContextInvocations++
	d.DownloadWithContextInputs = append(d.DownloadWithContextInputs, DownloadWithContextInput{WriterAt: writerAt, Input: input, Options: options})
	if d.DownloadWithContextStub != nil {
		return d.DownloadWithContextStub(ctx, writerAt, input, options...)
	}
	if len(d.DownloadWithContextOutputs) > 0 {
		output := d.DownloadWithContextOutputs[0]
		d.DownloadWithContextOutputs = d.DownloadWithContextOutputs[1:]
		return output.BytesWritten, output.Error
	}
	if d.DownloadWithContextOutput != nil {
		return d.DownloadWithContextOutput.BytesWritten, d.DownloadWithContextOutput.Error
	}
	panic("DownloadWithContext has no output")
}

func (d *Downloader) SetDownloadWithContextOutput(output DownloadWithContextOutput) {
	d.DownloadWithContextOutput = &output
}

func (d *Downloader) AssertOutputsEmpty() {
	if len(d.DownloadWithContextOutputs) > 0 {
		panic("DownloadWithContextOutputs is not empty")
	}
}
