package test

import (
	ouraClient "github.com/tidepool-org/platform/oura/client"
	"github.com/tidepool-org/platform/test"
)

func RandomErrorResponseDetail(options ...test.Option) ouraClient.ErrorResponseDetail {
	return ouraClient.ErrorResponseDetail{
		Location: test.RandomStringArray(),
		Message:  test.RandomString(),
		Type:     test.RandomString(),
	}
}

func RandomErrorResponse(options ...test.Option) ouraClient.ErrorResponse {
	detail := make([]ouraClient.ErrorResponseDetail, test.RandomIntFromRange(1, 5))
	for index := range detail {
		detail[index] = RandomErrorResponseDetail(options...)
	}
	return ouraClient.ErrorResponse{Detail: detail}
}
