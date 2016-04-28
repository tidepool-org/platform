package dataservices_test

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = PDescribe("The Dataservices client", func() {
	// var client *dataservices.DataServiceClient
	// var env map[string]interface{}
	// var params map[string]string
	// const userID = "9999999"
	// const groupID = "223377628"

	// BeforeEach(func() {
	// 	client = dataservices.NewDataServiceClient()
	// })

	// Describe("Version", func() {

	// 	BeforeEach(func() {
	// 		env = make(map[string]interface{})
	// 		params = make(map[string]string)
	// 	})

	// 	It("returns status 200", func() {
	// 		recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil), params, env)
	// 		Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
	// 		Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), "Expected "+string(recorded.Recorder.Code)+" to be "+string(http.StatusOK))
	// 	})
	// 	It("contains the version as the body", func() {
	// 		recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil), params, env)
	// 		Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
	// 		Expect(recorded.BodyIs(version.Long())).To(BeTrue(), "Expected "+recorded.Recorder.Body.String()+" to be "+version.Long())
	// 	})
	// 	It("has content type of json", func() {
	// 		recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil), params, env)
	// 		Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
	// 	})
	// })

	// Describe("PostDataset", func() {

	// 	const userID = "9999999"
	// 	var errorPayload interface{}

	// 	BeforeEach(func() {
	// 		env = make(map[string]interface{})
	// 		env[user.PERMISSIONS] = &user.UsersPermissions{}
	// 		env[user.GROUPID] = groupID

	// 		//the userid is used in the saving of the data so we attach it to the request in the `RunRequest` test handler
	// 		params = make(map[string]string)
	// 		params["userid"] = userID
	// 	})

	// 	Describe("when given valid data", func() {

	// 		jsonData := []byte(`[{"deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]`)

	// 		It("returns status 200", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
	// 		})

	// 		It("has content type of json", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
	// 		})

	// 		It("no data is returned", func() {
	// 			var payload interface{}
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			recorded.DecodeJSONPayload(&payload)
	// 			Expect(payload).To(BeNil(), "Expected that not data was returned")
	// 		})
	// 	})

	// 	Describe("when given valid data but wrong type", func() {

	// 		jsonData := []byte(`[{"userID": "9999999", "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "NOT_VALID", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]`)

	// 		It("returns status 400", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			Expect(recorded.CodeIs(http.StatusBadRequest)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusBadRequest))
	// 		})

	// 		It("has content type of json", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
	// 		})

	// 		It("should return an error with the payload", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			recorded.DecodeJSONPayload(&errorPayload)
	// 			Expect(errorPayload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 		})

	// 	})

	// 	Describe("when given invalid data", func() {

	// 		jsonData := []byte(`[{"blah": "9999999", "time": "2014-06-11T06:00:00.000Z"}]`)

	// 		It("returns status 400", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			Expect(recorded.CodeIs(http.StatusBadRequest)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusBadRequest))
	// 		})

	// 		It("has content type of json", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
	// 		})

	// 		It("should return an error saying there is no match for the type", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			recorded.DecodeJSONPayload(&errorPayload)
	// 			Expect(errorPayload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 		})

	// 	})

	// 	Describe("when any datapoint is invalid", func() {

	// 		//contains `"deliveryType": "unknown"` which does not pass validation
	// 		jsonData := []byte(`[{"deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"},{"deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "unknown", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]`)

	// 		It("returns status 400", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			Expect(recorded.CodeIs(http.StatusBadRequest)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusBadRequest))
	// 		})

	// 		It("should return an error saying there is no match for the type", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, bytes.NewBuffer(jsonData)), params, env)
	// 			recorded.DecodeJSONPayload(&errorPayload)
	// 			Expect(errorPayload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 		})

	// 	})

	// 	Describe("when given no data", func() {

	// 		It("returns status 400", func() {
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, nil), params, env)
	// 			Expect(recorded.CodeIs(http.StatusBadRequest)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusBadRequest))
	// 		})

	// 		It("should return body with error message", func() {
	// 			ExpectedError := `{"Error":"missing data to process"}`
	// 			recorded := service.RunRequest(client.PostDataset, service.MakeSimpleRequest("POST", "http://localhost/dataset/"+userID, nil), params, env)
	// 			Expect(recorded.BodyIs(ExpectedError)).To(BeTrue(), "Expected "+recorded.Recorder.Body.String()+" to be "+ExpectedError)
	// 		})

	// 	})

	// })
	// Describe("PostBlob", func() {

	// 	const userID = "9999999"
	// 	var fileName string

	// 	BeforeEach(func() {
	// 		env = make(map[string]interface{})
	// 		env[user.PERMISSIONS] = &user.UsersPermissions{}
	// 		env[user.GROUPID] = "3887276s"
	// 		fileName = ""
	// 	})

	// 	Describe("when given valid data", func() {

	// 		It("returns status 501", func() {
	// 			recorded := service.RunRequest(client.PostBlob, service.MakeBlobRequest("POST", "http://localhost/blob/"+userID, fileName), params, env)
	// 			Expect(recorded.CodeIs(http.StatusNotImplemented)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusNotImplemented))
	// 		})

	// 		It("should return be content type of json", func() {
	// 			recorded := service.RunRequest(client.PostBlob, service.MakeBlobRequest("POST", "http://localhost/blob/"+userID, fileName), params, env)
	// 			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
	// 		})
	// 	})

	// })
	// Describe("GetDataset", func() {

	// 	var payload struct {
	// 		Dataset []interface{} `json:"Dataset"`
	// 		Errors  string        `json:"Errors"`
	// 	}

	// 	var idParams map[string]string

	// 	BeforeEach(func() {
	// 		env = make(map[string]interface{})
	// 		env[user.PERMISSIONS] = &user.UsersPermissions{}
	// 	})

	// 	Describe("when valid userID", func() {

	// 		BeforeEach(func() {
	// 			idParams = make(map[string]string)
	// 			idParams["userid"] = userID
	// 			env[user.GROUPID] = groupID
	// 		})

	// 		It("returns status 200", func() {
	// 			recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
	// 			Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
	// 		})

	// 		It("has content type of json", func() {
	// 			recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
	// 			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
	// 		})

	// 		It("should return no error with the payload", func() {
	// 			recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
	// 			recorded.DecodeJSONPayload(&payload)
	// 			Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 			Expect(payload.Errors).To(Equal(""), "Expected the return errors to be empty")
	// 		})

	// 		It("should return the processed dataset with the payload", func() {
	// 			recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
	// 			recorded.DecodeJSONPayload(&payload)
	// 			Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 			Expect(len(payload.Dataset) > 1).To(BeTrue(), "Expected one processed datum to be returned")
	// 		})
	// 	})

	// 	Describe("when query params provided", func() {

	// 		var queryParams map[string]string

	// 		BeforeEach(func() {
	// 			queryParams = make(map[string]string)
	// 			queryParams["userid"] = userID
	// 			env[user.GROUPID] = groupID
	// 		})

	// 		It("should return basals", func() {

	// 			req := service.MakeSimpleRequest("GET", "http://localhost/dataset", nil)
	// 			url, _ := url.Parse("?type=basal")
	// 			req.URL = url

	// 			recorded := service.RunRequest(client.GetDataset, req, queryParams, env)
	// 			recorded.DecodeJSONPayload(&payload)
	// 			Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 			Expect(len(payload.Dataset) > 1).To(BeTrue(), "Expected one processed datum to be returned")
	// 		})
	// 		It("should return no basals when no subtype match", func() {
	// 			req := service.MakeSimpleRequest("GET", "http://localhost/dataset", nil)
	// 			url, _ := url.Parse("?type=basal&subType=good")
	// 			req.URL = url

	// 			recorded := service.RunRequest(client.GetDataset, req, queryParams, env)
	// 			recorded.DecodeJSONPayload(&payload)
	// 			Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 			Expect(len(payload.Dataset) == 0).To(BeTrue(), "Expected no processed datum to be returned")
	// 		})
	// 		It("should return nothing", func() {
	// 			req := service.MakeSimpleRequest("GET", "http://localhost/dataset", nil)
	// 			url, _ := url.Parse("?type=none")
	// 			req.URL = url

	// 			recorded := service.RunRequest(client.GetDataset, req, queryParams, env)
	// 			recorded.DecodeJSONPayload(&payload)
	// 			Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 			Expect(len(payload.Dataset) == 0).To(BeTrue(), "Expected no processed datum to be returned when no type match")
	// 		})
	// 	})

	// 	Describe("when userID unknown", func() {

	// 		const userID = "9???9"

	// 		BeforeEach(func() {
	// 			idParams = make(map[string]string)
	// 			idParams["userid"] = userID
	// 			env[user.GROUPID] = userID
	// 		})

	// 		It("returns status 200", func() {
	// 			recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
	// 			Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), fmt.Sprintf("Expected %d to be %d", recorded.Recorder.Code, http.StatusOK))
	// 		})

	// 		It("has content type of json", func() {
	// 			recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
	// 			Expect(recorded.ContentTypeIsJSON()).To(BeTrue(), "Expected content type to be JSON")
	// 		})

	// 		It("should return no error with the payload", func() {
	// 			recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
	// 			recorded.DecodeJSONPayload(&payload)
	// 			Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 			Expect(payload.Errors).To(Equal(""), "Expected the return errors to be empty")
	// 		})

	// 		It("should return no dataset with the payload", func() {
	// 			recorded := service.RunRequest(client.GetDataset, service.MakeSimpleRequest("GET", "http://localhost/dataset/"+userID, nil), idParams, env)
	// 			recorded.DecodeJSONPayload(&payload)
	// 			Expect(payload).ToNot(BeNil(), "Expected the return payload to not be nil")
	// 			Expect(len(payload.Dataset)).To(Equal(0), "Expected no data to be returned")
	// 		})

	// 	})
	// })

})
