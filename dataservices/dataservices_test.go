package main_test

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	. "github.com/tidepool-org/platform/dataservices"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("The Dataservices client", func() {
	var client *DataServiceClient
	var env map[string]interface{}
	var params map[string]string

	BeforeEach(func() {
		client = NewDataServiceClient()
		env = make(map[string]interface{})
		params = make(map[string]string)
	})

	AfterEach(func() {
		//shut down the server between tests
	})

	Describe("Version", func() {
		It("should return status 200", func() {
			recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil), params, env)
			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), "Expected "+string(recorded.Recorder.Code)+" to be "+string(http.StatusOK))
		})
		It("should return version as the body", func() {
			recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil), params, env)
			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			Expect(recorded.BodyIs(version.String)).To(BeTrue(), "Expected "+recorded.Recorder.Body.String()+" to be "+version.String)
		})
		It("should be content type of json", func() {
			recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil), params, env)
			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
		})
	})

	Describe("PostDataset", func() {

		const userId = "9999999"
		var payload struct {
			Dataset []interface{} `json:"Dataset"`
			Errors  string        `json:"Errors"`
		}

		perms := make(map[string]interface{})
		perms[user.PERMISSIONS] = &user.UsersPermissions{}

		//the userid is used in the saving of the data so we attach it to the request in the `RunRequest` test handler
		params := make(map[string]string)
		params["userid"] = userId

		Describe("when given valid data", func() {

			jsonData := []byte(`[{"deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]`)

			It("should return status 200", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return no error with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(payload.Errors).To(Equal(""), "Expected the return errors to be empty")
			})

			It("should return the processed dataset with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset)).To(Equal(1), "Expected one processed datum to be returned")
			})
		})

		Describe("when given valid data but wrong type", func() {

			jsonData := []byte(`[{"userId": "9999999", "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "NOT_VALID", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]`)

			It("should return status 200", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return an error with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(payload.Errors).ToNot(Equal(""), "Expected the return errors to not be empty")
				Expect(strings.Contains(payload.Errors, "we can't deal with `type`=NOT_VALID")).To(BeTrue(), "Expected the return errors to not be empty")
			})

			It("should return the no items in the processed dataset with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset)).To(Equal(0), "Expected no processed datum to be returned")
			})

		})

		Describe("when given invalid data", func() {

			jsonData := []byte(`[{"blah": "9999999", "time": "2014-06-11T06:00:00.000Z"}]`)

			It("should return status 200", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return an error saying there is no match for the type", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(strings.Contains(payload.Errors, "there is no match for that type")).To(BeTrue(), "Expected the return errors to not be empty")
			})

			It("should return the no items in the processed dataset with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, bytes.NewBuffer(jsonData)), params, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset)).To(Equal(0), "Expected no processed datum to be returned")
			})

		})

		Describe("when given no data", func() {

			It("should return status 400", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, nil), params, perms)
				Expect(recorded.CodeIs(http.StatusBadRequest)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusBadRequest))
			})

			It("should return body with error message", func() {
				expectedError := `{"Error":"missing data to process"}`
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userId, nil), params, perms)
				Expect(recorded.BodyIs(expectedError)).To(BeTrue(), "Expected "+recorded.Recorder.Body.String()+" to be "+expectedError)
			})

		})

	})
	Describe("PostBlob", func() {

		const userId = "9999999"

		perms := make(map[string]interface{})
		perms[user.PERMISSIONS] = &user.UsersPermissions{}
		fileName := ""

		Describe("when given valid data", func() {

			It("should return status 200", func() {
				recorded := service.RunRequest(client.PostBlob, service.MakeBlobRequest("POST", "http://localhost/blob/"+userId, fileName), params, perms)
				Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
			})

			It("should return be content type of json", func() {
				recorded := service.RunRequest(client.PostBlob, service.MakeBlobRequest("POST", "http://localhost/blob/"+userId, fileName), params, perms)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})
		})

	})
	Describe("GetDataset", func() {

		var payload struct {
			Dataset []interface{} `json:"Dataset"`
			Errors  string        `json:"Errors"`
		}

		perms := make(map[string]interface{})
		perms[user.PERMISSIONS] = &user.UsersPermissions{}

		Describe("when valid userId", func() {

			const userId = "9999999"
			idParams := make(map[string]string)
			idParams["userid"] = userId

			It("should return status 200", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userId, nil), idParams, perms)
				Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userId, nil), idParams, perms)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return no error with the payload", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userId, nil), idParams, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(payload.Errors).To(Equal(""), "Expected the return errors to be empty")
			})

			It("should return the processed dataset with the payload", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userId, nil), idParams, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset) > 1).To(BeTrue(), "Expected one processed datum to be returned")
			})
		})

		Describe("when userId unknown", func() {

			const userId = "9???9"
			idParams := make(map[string]string)
			idParams["userid"] = userId

			It("should return status 200", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userId, nil), idParams, perms)
				Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userId, nil), idParams, perms)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return no error with the payload", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userId, nil), idParams, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(payload.Errors).To(Equal(""), "Expected the return errors to be empty")
			})

			It("should return no dataset with the payload", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userId, nil), idParams, perms)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset)).To(Equal(0), "Expected no data to be returned")
			})

		})
	})

})
