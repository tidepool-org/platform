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

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("The Dataservices client", func() {
	var client *DataServiceClient
	var env map[string]interface{}
	var params map[string]string

	BeforeEach(func() {
		client = NewDataServiceClient()
	})

	Describe("Version", func() {

		BeforeEach(func() {
			env = make(map[string]interface{})
			params = make(map[string]string)
		})

		It("should return status 200", func() {
			recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil), params, env)
			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), "Expected "+string(recorded.Recorder.Code)+" to be "+string(http.StatusOK))
		})
		It("should return version as the body", func() {
			recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil), params, env)
			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			Expect(recorded.BodyIs(version.Long())).To(BeTrue(), "Expected "+recorded.Recorder.Body.String()+" to be "+version.Long())
		})
		It("should be content type of json", func() {
			recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil), params, env)
			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
		})
	})

	Describe("PostDataset", func() {

		const userID = "9999999"
		var payload struct {
			Dataset []interface{} `json:"Dataset"`
			Errors  string        `json:"Errors"`
		}

		BeforeEach(func() {
			env = make(map[string]interface{})
			env[user.PERMISSIONS] = &user.UsersPermissions{}
			env[user.GROUPID] = "223377628"

			//the userid is used in the saving of the data so we attach it to the request in the `RunRequest` test handler
			params = make(map[string]string)
			params["userid"] = userID
		})

		Describe("when given valid data", func() {

			jsonData := []byte(`[{"deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]`)

			It("should return status 200", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return no error with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(payload.Errors).To(Equal(""), "Expected the return errors to be empty")
			})

			It("should return the processed dataset with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset)).To(Equal(1), "Expected one processed datum to be returned")
			})
		})

		Describe("when given valid data but wrong type", func() {

			jsonData := []byte(`[{"userID": "9999999", "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "NOT_VALID", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]`)

			It("should return status 400", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				Expect(recorded.CodeIs(http.StatusBadRequest)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusBadRequest))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return an error with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(payload.Errors).ToNot(Equal(""), "Expected the return errors to not be empty")
				Expect(strings.Contains(payload.Errors, "we can't deal with `type`=NOT_VALID")).To(BeTrue(), "Expected the return errors to not be empty")
			})

			It("should return the no items in the processed dataset with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset)).To(Equal(0), "Expected no processed datum to be returned")
			})

		})

		Describe("when given invalid data", func() {

			jsonData := []byte(`[{"blah": "9999999", "time": "2014-06-11T06:00:00.000Z"}]`)

			It("should return status 400", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				Expect(recorded.CodeIs(http.StatusBadRequest)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusBadRequest))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return an error saying there is no match for the type", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(strings.Contains(payload.Errors, "there is no match for that type")).To(BeTrue(), "Expected the return errors to not be empty")
			})

			It("should return the no items in the processed dataset with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset)).To(Equal(0), "Expected no processed datum to be returned")
			})

		})

		Describe("when any datapoint is invalid", func() {

			//contains `"deliveryType": "unknown"` which does not pass validation
			jsonData := []byte(`[{"deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"},{"deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "unknown", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]`)

			It("should return status 400", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				Expect(recorded.CodeIs(http.StatusBadRequest)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusBadRequest))
			})

			It("should return an error saying there is no match for the type", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(payload.Errors).ToNot(BeZero(), "Expected the payload error to be set")
			})

			It("should return the valid items in the processed dataset with the payload", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset)).To(Equal(1), "Expected one processed datum to be returned")
			})

		})

		Describe("when given no data", func() {

			It("should return status 400", func() {
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, nil), params, env)
				Expect(recorded.CodeIs(http.StatusBadRequest)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusBadRequest))
			})

			It("should return body with error message", func() {
				expectedError := `{"Error":"missing data to process"}`
				recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, nil), params, env)
				Expect(recorded.BodyIs(expectedError)).To(BeTrue(), "Expected "+recorded.Recorder.Body.String()+" to be "+expectedError)
			})

		})

	})
	Describe("PostBlob", func() {

		const userID = "9999999"
		var fileName string

		BeforeEach(func() {
			env = make(map[string]interface{})
			env[user.PERMISSIONS] = &user.UsersPermissions{}
			env[user.GROUPID] = "3887276s"
			fileName = ""
		})

		Describe("when given valid data", func() {

			It("should return status 501", func() {
				recorded := service.RunRequest(client.PostBlob, service.MakeBlobRequest("POST", "http://localhost/blob/"+userID, fileName), params, env)
				Expect(recorded.CodeIs(http.StatusNotImplemented)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusNotImplemented))
			})

			It("should return be content type of json", func() {
				recorded := service.RunRequest(client.PostBlob, service.MakeBlobRequest("POST", "http://localhost/blob/"+userID, fileName), params, env)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})
		})

	})
	Describe("GetDataset", func() {

		var payload struct {
			Dataset []interface{} `json:"Dataset"`
			Errors  string        `json:"Errors"`
		}

		var idParams map[string]string

		BeforeEach(func() {
			env = make(map[string]interface{})
			env[user.PERMISSIONS] = &user.UsersPermissions{}
			env[user.GROUPID] = "3887276s"
		})

		Describe("when valid userID", func() {

			const userID = "9999999"
			BeforeEach(func() {
				idParams = make(map[string]string)
				idParams["userid"] = userID
			})

			It("should return status 200", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
				Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return no error with the payload", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(payload.Errors).To(Equal(""), "Expected the return errors to be empty")
			})

			It("should return the processed dataset with the payload", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset) > 1).To(BeTrue(), "Expected one processed datum to be returned")
			})
		})

		Describe("when userID unknown", func() {

			const userID = "9???9"
			BeforeEach(func() {
				idParams = make(map[string]string)
				idParams["userid"] = userID
			})

			It("should return status 200", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
				Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
			})

			It("should be content type of json", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
				Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
			})

			It("should return no error with the payload", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(payload.Errors).To(Equal(""), "Expected the return errors to be empty")
			})

			It("should return no dataset with the payload", func() {
				recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
				recorded.DecodeJSONPayload(&payload)
				Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
				Expect(len(payload.Dataset)).To(Equal(0), "Expected no data to be returned")
			})

		})
	})

})
