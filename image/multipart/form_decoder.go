package multipart

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/image"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type FormDecoder interface {
	DecodeForm(reader io.Reader, contentType string) (*image.Metadata, string, *image.Content, error)
}

type FormDecoderImpl struct{}

func NewFormDecoder() *FormDecoderImpl {
	return &FormDecoderImpl{}
}

func (f *FormDecoderImpl) DecodeForm(reader io.Reader, contentType string) (*image.Metadata, string, *image.Content, error) {
	if reader == nil {
		return nil, "", nil, errors.New("reader is missing")
	}
	if contentType == "" {
		return nil, "", nil, errors.New("content type is missing")
	}

	var boundary string
	if mediaType, parameters, err := mime.ParseMediaType(contentType); err != nil {
		return nil, "", nil, errors.New("content type is invalid")
	} else if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, "", nil, errors.New("content type is not supported")
	} else if boundary = parameters["boundary"]; boundary == "" {
		return nil, "", nil, errors.New("boundary is missing")
	}

	form, err := multipart.NewReader(reader, boundary).ReadForm(decodeFormMemoryMaximum)
	if err != nil {
		return nil, "", nil, errors.Wrap(err, "unable to create multipart reader")
	}
	defer func() {
		if form != nil {
			form.RemoveAll()
		}
	}()

	var metadata *image.Metadata
	var contentIntent string
	var content *image.Content

	var foundMetadata bool
	var foundContentIntent bool
	for key, values := range form.Value {
		keySource := structure.NewPointerSource().WithReference(key)
		for index, value := range values {
			valueSource := keySource.WithReference(strconv.Itoa(index))
			decodeErr := errors.WithSource(structureParser.ErrorNotParsed(), valueSource)
			if index == 0 {
				switch key {
				case "metadata":
					foundMetadata = true
					metadata = image.NewMetadata()
					decodeErr = request.DecodeObject(valueSource, strings.NewReader(value), metadata)
				case "contentIntent":
					foundContentIntent = true
					if !image.IsValidContentIntent(value) {
						decodeErr = errors.WithSource(structureValidator.ErrorValueStringNotOneOf(value, image.ContentIntents()), valueSource)
					} else {
						contentIntent = value
						decodeErr = nil
					}
				}
			}
			if decodeErr != nil {
				err = errors.Append(err, decodeErr)
			}
		}
	}
	if !foundMetadata {
		metadata = image.NewMetadata()
	}
	if !foundContentIntent {
		err = errors.Append(err, errors.WithSource(structureValidator.ErrorValueNotExists(), structure.NewPointerSource().WithReference("contentIntent")))
	}

	var foundContent bool
	for key, values := range form.File {
		keySource := structure.NewPointerSource().WithReference(key)
		for index, value := range values {
			valueSource := keySource.WithReference(strconv.Itoa(index))
			decodeErr := structureParser.ErrorNotParsed()
			if index == 0 {
				switch key {
				case "content":
					foundContent = true
					var name *string
					content, name, decodeErr = f.readFormContent(value)
					if name != nil && metadata.Name == nil {
						metadata.Name = name
					}
				}
			}
			if decodeErr != nil {
				err = errors.Append(err, errors.WithSource(decodeErr, valueSource))
			}
		}
	}
	if !foundContent {
		err = errors.Append(err, errors.WithSource(structureValidator.ErrorValueNotExists(), structure.NewPointerSource().WithReference("content")))
	}

	if err != nil {
		if content != nil && content.Body != nil {
			content.Body.Close()
		}
		return nil, "", nil, err
	}

	content.Body = &formBody{Form: form, Body: content.Body}
	form = nil

	return metadata, contentIntent, content, nil
}

func (f *FormDecoderImpl) readFormContent(multipartFileHeader *multipart.FileHeader) (*image.Content, *string, error) {
	var name *string
	if filename := strings.TrimSpace(multipartFileHeader.Filename); filename != "" {
		name = pointer.FromString(filename)
	}

	header := http.Header(multipartFileHeader.Header)
	mediaType, err := request.ParseMediaTypeHeader(header, "Content-Type")
	if err != nil {
		return nil, nil, err
	} else if mediaType == nil {
		return nil, nil, request.ErrorHeaderMissing("Content-Type")
	} else if err = image.ValidateMediaType(*mediaType); err != nil {
		return nil, nil, err
	}
	digestMD5, err := request.ParseDigestMD5Header(header, "Digest")
	if err != nil {
		return nil, nil, err
	}

	body, err := multipartFileHeader.Open()
	if err != nil {
		return nil, nil, err
	}

	content := image.NewContent()
	content.Body = body
	content.DigestMD5 = digestMD5
	content.MediaType = mediaType
	content.DigestMD5 = digestMD5
	return content, name, nil
}

type formBody struct {
	Form *multipart.Form
	Body io.ReadCloser
}

func (f *formBody) Read(bytes []byte) (int, error) {
	return f.Body.Read(bytes)
}

func (f *formBody) Close() error {
	err := f.Body.Close()
	if formErr := f.Form.RemoveAll(); err == nil {
		err = formErr
	}
	return err
}

const decodeFormMemoryMaximum = 10 * 1024 * 1024
