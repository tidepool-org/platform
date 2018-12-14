package multipart

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"

	"github.com/tidepool-org/platform/image"
)

type FormEncoder interface {
	EncodeForm(metadata *image.Metadata, contentIntent string, content *image.Content) (io.ReadCloser, string)
}

type FormEncoderImpl struct{}

func NewFormEncoder() *FormEncoderImpl {
	return &FormEncoderImpl{}
}

func (f *FormEncoderImpl) EncodeForm(metadata *image.Metadata, contentIntent string, content *image.Content) (io.ReadCloser, string) {
	pipeReader, pipeWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(pipeWriter)
	contentType := multipartWriter.FormDataContentType()
	go func() {
		var err error

		defer func() {
			if closeErr := multipartWriter.Close(); err == nil {
				err = closeErr
			}
			pipeWriter.CloseWithError(err)
		}()

		err = f.writeForm(metadata, contentIntent, content, multipartWriter)
	}()

	return pipeReader, contentType
}

func (f *FormEncoderImpl) writeForm(metadata *image.Metadata, contentIntent string, content *image.Content, multipartWriter *multipart.Writer) error {
	if metadata != nil {
		if metadataBytes, err := json.Marshal(metadata); err != nil {
			return err
		} else if err = multipartWriter.WriteField("metadata", string(metadataBytes)); err != nil {
			return err
		}
	}

	if contentIntent != "" {
		if err := multipartWriter.WriteField("contentIntent", contentIntent); err != nil {
			return err
		}
	}

	if content != nil && content.Body != nil {
		contentHeader := textproto.MIMEHeader{}
		contentHeader.Set("Content-Disposition", `form-data; name="content"; filename=" "`)
		if content.MediaType != nil {
			contentHeader.Set("Content-Type", *content.MediaType)
		}
		if content.DigestMD5 != nil {
			contentHeader.Set("Digest", fmt.Sprintf("MD5=%s", *content.DigestMD5))
		}
		if contentPart, err := multipartWriter.CreatePart(contentHeader); err != nil {
			return err
		} else if _, err = io.Copy(contentPart, content.Body); err != nil {
			return err
		}
	}
	return nil
}
