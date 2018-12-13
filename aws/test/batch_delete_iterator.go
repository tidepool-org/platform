package test

import awsSdkGoServiceS3S3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"

type BatchDeleteIterator struct {
	NextInvocations         int
	NextStub                func() bool
	NextOutputs             []bool
	NextOutput              *bool
	ErrInvocations          int
	ErrStub                 func() error
	ErrOutputs              []error
	ErrOutput               *error
	DeleteObjectInvocations int
	DeleteObjectStub        func() awsSdkGoServiceS3S3manager.BatchDeleteObject
	DeleteObjectOutputs     []awsSdkGoServiceS3S3manager.BatchDeleteObject
	DeleteObjectOutput      *awsSdkGoServiceS3S3manager.BatchDeleteObject
}

func NewBatchDeleteIterator() *BatchDeleteIterator {
	return &BatchDeleteIterator{}
}

func (b *BatchDeleteIterator) Next() bool {
	b.NextInvocations++
	if b.NextStub != nil {
		return b.NextStub()
	}
	if len(b.NextOutputs) > 0 {
		output := b.NextOutputs[0]
		b.NextOutputs = b.NextOutputs[1:]
		return output
	}
	if b.NextOutput != nil {
		return *b.NextOutput
	}
	panic("Next has no output")
}

func (b *BatchDeleteIterator) Err() error {
	b.ErrInvocations++
	if b.ErrStub != nil {
		return b.ErrStub()
	}
	if len(b.ErrOutputs) > 0 {
		output := b.ErrOutputs[0]
		b.ErrOutputs = b.ErrOutputs[1:]
		return output
	}
	if b.ErrOutput != nil {
		return *b.ErrOutput
	}
	panic("Err has no output")
}

func (b *BatchDeleteIterator) DeleteObject() awsSdkGoServiceS3S3manager.BatchDeleteObject {
	b.DeleteObjectInvocations++
	if b.DeleteObjectStub != nil {
		return b.DeleteObjectStub()
	}
	if len(b.DeleteObjectOutputs) > 0 {
		output := b.DeleteObjectOutputs[0]
		b.DeleteObjectOutputs = b.DeleteObjectOutputs[1:]
		return output
	}
	if b.DeleteObjectOutput != nil {
		return *b.DeleteObjectOutput
	}
	panic("DeleteObject has no output")
}

func (b *BatchDeleteIterator) SetNextOutput(output bool) {
	b.NextOutput = &output
}

func (b *BatchDeleteIterator) SetErrOutput(output error) {
	b.ErrOutput = &output
}

func (b *BatchDeleteIterator) SetDeleteObjectOutput(output awsSdkGoServiceS3S3manager.BatchDeleteObject) {
	b.DeleteObjectOutput = &output
}

func (b *BatchDeleteIterator) AssertOutputsEmpty() {
	if len(b.NextOutputs) > 0 {
		panic("NextOutputs is not empty")
	}
	if len(b.ErrOutputs) > 0 {
		panic("ErrOutputs is not empty")
	}
	if len(b.DeleteObjectOutputs) > 0 {
		panic("DeleteObjectOutputs is not empty")
	}
}
