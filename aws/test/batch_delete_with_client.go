package test

import (
	awsSdkGoAws "github.com/aws/aws-sdk-go/aws"
	awsSdkGoServiceS3S3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type BatchDeleteWithClient struct {
	DeleteInvocations int
	DeleteInputs      []awsSdkGoServiceS3S3manager.BatchDeleteIterator
	DeleteStub        func(ctx awsSdkGoAws.Context, batchDeleteIterator awsSdkGoServiceS3S3manager.BatchDeleteIterator) error
	DeleteOutputs     []error
	DeleteOutput      *error
}

func NewBatchDeleteWithClient() *BatchDeleteWithClient {
	return &BatchDeleteWithClient{}
}

func (b *BatchDeleteWithClient) Delete(ctx awsSdkGoAws.Context, batchDeleteIterator awsSdkGoServiceS3S3manager.BatchDeleteIterator) error {
	b.DeleteInvocations++
	b.DeleteInputs = append(b.DeleteInputs, batchDeleteIterator)
	if b.DeleteStub != nil {
		return b.DeleteStub(ctx, batchDeleteIterator)
	}
	if len(b.DeleteOutputs) > 0 {
		output := b.DeleteOutputs[0]
		b.DeleteOutputs = b.DeleteOutputs[1:]
		return output
	}
	if b.DeleteOutput != nil {
		return *b.DeleteOutput
	}
	panic("Delete has no output")
}

func (b *BatchDeleteWithClient) SetDeleteOutput(output error) {
	b.DeleteOutput = &output
}

func (b *BatchDeleteWithClient) AssertOutputsEmpty() {
	if len(b.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
}
