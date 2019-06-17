package test

import (
	"context"
	"io"

	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
)

type PutContentInput struct {
	UserID        string
	ImageID       string
	ContentID     string
	ContentIntent string
	Reader        io.Reader
	Options       *storeUnstructured.Options
}

type GetContentInput struct {
	UserID        string
	ImageID       string
	ContentID     string
	ContentIntent string
}

type GetContentOutput struct {
	Reader io.ReadCloser
	Error  error
}

type DeleteContentInput struct {
	UserID    string
	ImageID   string
	ContentID string
}

type PutRenditionContentInput struct {
	UserID       string
	ImageID      string
	ContentID    string
	RenditionsID string
	Rendition    string
	Reader       io.Reader
	Options      *storeUnstructured.Options
}

type GetRenditionContentInput struct {
	UserID       string
	ImageID      string
	ContentID    string
	RenditionsID string
	Rendition    string
}

type GetRenditionContentOutput struct {
	Reader io.ReadCloser
	Error  error
}

type DeleteRenditionContentInput struct {
	UserID       string
	ImageID      string
	ContentID    string
	RenditionsID string
}

type DeleteInput struct {
	UserID  string
	ImageID string
}

type Store struct {
	PutContentInvocations             int
	PutContentInputs                  []PutContentInput
	PutContentStub                    func(ctx context.Context, userID string, imageID string, contentID string, contentIntent string, reader io.Reader, options *storeUnstructured.Options) error
	PutContentOutputs                 []error
	PutContentOutput                  *error
	GetContentInvocations             int
	GetContentInputs                  []GetContentInput
	GetContentStub                    func(ctx context.Context, userID string, imageID string, contentID string, contentIntent string) (io.ReadCloser, error)
	GetContentOutputs                 []GetContentOutput
	GetContentOutput                  *GetContentOutput
	DeleteContentInvocations          int
	DeleteContentInputs               []DeleteContentInput
	DeleteContentStub                 func(ctx context.Context, userID string, imageID string, contentID string) error
	DeleteContentOutputs              []error
	DeleteContentOutput               *error
	PutRenditionContentInvocations    int
	PutRenditionContentInputs         []PutRenditionContentInput
	PutRenditionContentStub           func(ctx context.Context, userID string, imageID string, contentID string, renditionsID string, rendition string, reader io.Reader, options *storeUnstructured.Options) error
	PutRenditionContentOutputs        []error
	PutRenditionContentOutput         *error
	GetRenditionContentInvocations    int
	GetRenditionContentInputs         []GetRenditionContentInput
	GetRenditionContentStub           func(ctx context.Context, userID string, imageID string, contentID string, renditionsID string, rendition string) (io.ReadCloser, error)
	GetRenditionContentOutputs        []GetRenditionContentOutput
	GetRenditionContentOutput         *GetRenditionContentOutput
	DeleteRenditionContentInvocations int
	DeleteRenditionContentInputs      []DeleteRenditionContentInput
	DeleteRenditionContentStub        func(ctx context.Context, userID string, imageID string, contentID string, renditionsID string) error
	DeleteRenditionContentOutputs     []error
	DeleteRenditionContentOutput      *error
	DeleteInvocations                 int
	DeleteInputs                      []DeleteInput
	DeleteStub                        func(ctx context.Context, userID string, imageID string) error
	DeleteOutputs                     []error
	DeleteOutput                      *error
	DeleteAllInvocations              int
	DeleteAllInputs                   []string
	DeleteAllStub                     func(ctx context.Context, userID string) error
	DeleteAllOutputs                  []error
	DeleteAllOutput                   *error
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) PutContent(ctx context.Context, userID string, imageID string, contentID string, contentIntent string, reader io.Reader, options *storeUnstructured.Options) error {
	s.PutContentInvocations++
	s.PutContentInputs = append(s.PutContentInputs, PutContentInput{UserID: userID, ImageID: imageID, ContentID: contentID, ContentIntent: contentIntent, Reader: reader, Options: options})
	if s.PutContentStub != nil {
		return s.PutContentStub(ctx, userID, imageID, contentID, contentIntent, reader, options)
	}
	if len(s.PutContentOutputs) > 0 {
		output := s.PutContentOutputs[0]
		s.PutContentOutputs = s.PutContentOutputs[1:]
		return output
	}
	if s.PutContentOutput != nil {
		return *s.PutContentOutput
	}
	panic("PutContent has no output")
}

func (s *Store) GetContent(ctx context.Context, userID string, imageID string, contentID string, contentIntent string) (io.ReadCloser, error) {
	s.GetContentInvocations++
	s.GetContentInputs = append(s.GetContentInputs, GetContentInput{UserID: userID, ImageID: imageID, ContentID: contentID, ContentIntent: contentIntent})
	if s.GetContentStub != nil {
		return s.GetContentStub(ctx, userID, imageID, contentID, contentIntent)
	}
	if len(s.GetContentOutputs) > 0 {
		output := s.GetContentOutputs[0]
		s.GetContentOutputs = s.GetContentOutputs[1:]
		return output.Reader, output.Error
	}
	if s.GetContentOutput != nil {
		return s.GetContentOutput.Reader, s.GetContentOutput.Error
	}
	panic("GetContent has no output")
}

func (s *Store) DeleteContent(ctx context.Context, userID string, imageID string, contentID string) error {
	s.DeleteContentInvocations++
	s.DeleteContentInputs = append(s.DeleteContentInputs, DeleteContentInput{UserID: userID, ImageID: imageID, ContentID: contentID})
	if s.DeleteContentStub != nil {
		return s.DeleteContentStub(ctx, userID, imageID, contentID)
	}
	if len(s.DeleteContentOutputs) > 0 {
		output := s.DeleteContentOutputs[0]
		s.DeleteContentOutputs = s.DeleteContentOutputs[1:]
		return output
	}
	if s.DeleteContentOutput != nil {
		return *s.DeleteContentOutput
	}
	panic("DeleteContent has no output")
}

func (s *Store) PutRenditionContent(ctx context.Context, userID string, imageID string, contentID string, renditionsID string, rendition string, reader io.Reader, options *storeUnstructured.Options) error {
	s.PutRenditionContentInvocations++
	s.PutRenditionContentInputs = append(s.PutRenditionContentInputs, PutRenditionContentInput{UserID: userID, ImageID: imageID, ContentID: contentID, RenditionsID: renditionsID, Rendition: rendition, Reader: reader, Options: options})
	if s.PutRenditionContentStub != nil {
		return s.PutRenditionContentStub(ctx, userID, imageID, contentID, renditionsID, rendition, reader, options)
	}
	if len(s.PutRenditionContentOutputs) > 0 {
		output := s.PutRenditionContentOutputs[0]
		s.PutRenditionContentOutputs = s.PutRenditionContentOutputs[1:]
		return output
	}
	if s.PutRenditionContentOutput != nil {
		return *s.PutRenditionContentOutput
	}
	panic("PutRenditionContent has no output")
}

func (s *Store) GetRenditionContent(ctx context.Context, userID string, imageID string, contentID string, renditionsID string, rendition string) (io.ReadCloser, error) {
	s.GetRenditionContentInvocations++
	s.GetRenditionContentInputs = append(s.GetRenditionContentInputs, GetRenditionContentInput{UserID: userID, ImageID: imageID, ContentID: contentID, RenditionsID: renditionsID, Rendition: rendition})
	if s.GetRenditionContentStub != nil {
		return s.GetRenditionContentStub(ctx, userID, imageID, contentID, renditionsID, rendition)
	}
	if len(s.GetRenditionContentOutputs) > 0 {
		output := s.GetRenditionContentOutputs[0]
		s.GetRenditionContentOutputs = s.GetRenditionContentOutputs[1:]
		return output.Reader, output.Error
	}
	if s.GetRenditionContentOutput != nil {
		return s.GetRenditionContentOutput.Reader, s.GetRenditionContentOutput.Error
	}
	panic("GetRenditionContent has no output")
}

func (s *Store) DeleteRenditionContent(ctx context.Context, userID string, imageID string, contentID string, renditionsID string) error {
	s.DeleteRenditionContentInvocations++
	s.DeleteRenditionContentInputs = append(s.DeleteRenditionContentInputs, DeleteRenditionContentInput{UserID: userID, ImageID: imageID, ContentID: contentID, RenditionsID: renditionsID})
	if s.DeleteRenditionContentStub != nil {
		return s.DeleteRenditionContentStub(ctx, userID, imageID, contentID, renditionsID)
	}
	if len(s.DeleteRenditionContentOutputs) > 0 {
		output := s.DeleteRenditionContentOutputs[0]
		s.DeleteRenditionContentOutputs = s.DeleteRenditionContentOutputs[1:]
		return output
	}
	if s.DeleteRenditionContentOutput != nil {
		return *s.DeleteRenditionContentOutput
	}
	panic("DeleteRenditionContent has no output")
}

func (s *Store) Delete(ctx context.Context, userID string, imageID string) error {
	s.DeleteInvocations++
	s.DeleteInputs = append(s.DeleteInputs, DeleteInput{UserID: userID, ImageID: imageID})
	if s.DeleteStub != nil {
		return s.DeleteStub(ctx, userID, imageID)
	}
	if len(s.DeleteOutputs) > 0 {
		output := s.DeleteOutputs[0]
		s.DeleteOutputs = s.DeleteOutputs[1:]
		return output
	}
	if s.DeleteOutput != nil {
		return *s.DeleteOutput
	}
	panic("Delete has no output")
}

func (s *Store) DeleteAll(ctx context.Context, userID string) error {
	s.DeleteAllInvocations++
	s.DeleteAllInputs = append(s.DeleteAllInputs, userID)
	if s.DeleteAllStub != nil {
		return s.DeleteAllStub(ctx, userID)
	}
	if len(s.DeleteAllOutputs) > 0 {
		output := s.DeleteAllOutputs[0]
		s.DeleteAllOutputs = s.DeleteAllOutputs[1:]
		return output
	}
	if s.DeleteAllOutput != nil {
		return *s.DeleteAllOutput
	}
	panic("DeleteAll has no output")
}

func (s *Store) AssertOutputsEmpty() {
	if len(s.PutContentOutputs) > 0 {
		panic("PutContentOutputs is not empty")
	}
	if len(s.GetContentOutputs) > 0 {
		panic("GetContentOutputs is not empty")
	}
	if len(s.DeleteContentOutputs) > 0 {
		panic("DeleteContentOutputs is not empty")
	}
	if len(s.PutRenditionContentOutputs) > 0 {
		panic("PutRenditionContentOutputs is not empty")
	}
	if len(s.GetRenditionContentOutputs) > 0 {
		panic("GetRenditionContentOutputs is not empty")
	}
	if len(s.DeleteRenditionContentOutputs) > 0 {
		panic("DeleteRenditionContentOutputs is not empty")
	}
	if len(s.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
	if len(s.DeleteAllOutputs) > 0 {
		panic("DeleteAllOutputs is not empty")
	}
}
