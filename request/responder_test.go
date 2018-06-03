package request_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHTTP "github.com/tidepool-org/platform/test/http"
	testRest "github.com/tidepool-org/platform/test/rest"
)

type Data struct {
	Value         interface{} `json:"value"`
	SanitizeError error       `json:"-"`
}

func NewData() *Data {
	return &Data{
		Value: testHTTP.NewHeaderValue(),
	}
}

func (d *Data) Sanitize(details request.Details) error {
	return d.SanitizeError
}

type Error struct {
	Value interface{} `json:"value"`
}

func NewError() *Error {
	return &Error{
		Value: testHTTP.NewHeaderValue(),
	}
}

func (e *Error) Error() string {
	return "error"
}

var _ = Describe("Responder", func() {
	var res *testRest.ResponseWriter
	var req *rest.Request

	BeforeEach(func() {
		res = testRest.NewResponseWriter()
		req = testRest.NewRequest()
		req.Method = "POST"
		req.Request = req.WithContext(log.NewContextWithLogger(req.Context(), logNull.NewLogger()))
	})

	AfterEach(func() {
		res.Expectations()
	})

	Context("MustNewResponder", func() {
		It("panics if the response is missing", func() {
			Expect(func() { request.MustNewResponder(nil, req) }).To(Panic())
		})

		It("panics if the response is missing", func() {
			Expect(func() { request.MustNewResponder(res, nil) }).To(Panic())
		})

		It("succeeds", func() {
			Expect(request.MustNewResponder(res, req)).ToNot(BeNil())
		})
	})

	Context("NewResponder", func() {
		It("returns an error if the response is missing", func() {
			responder, err := request.NewResponder(nil, req)
			Expect(err).To(MatchError("response is missing"))
			Expect(responder).To(BeNil())
		})

		It("returns an error if the request is missing", func() {
			responder, err := request.NewResponder(res, nil)
			Expect(err).To(MatchError("request is missing"))
			Expect(responder).To(BeNil())
		})

		It("succeeds", func() {
			Expect(request.NewResponder(res, req)).ToNot(BeNil())
		})
	})

	Context("with new responder", func() {
		var responder *request.Responder

		BeforeEach(func() {
			var err error
			responder, err = request.NewResponder(res, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(responder).ToNot(BeNil())
		})

		Context("SetCookie", func() {
			It("does nothing if cookie is nil", func() {
				responder.SetCookie(nil)
				Expect(res.HeaderImpl).To(BeEmpty())
			})

			It("adds the expected header if the cookie is not nil", func() {
				cookie := http.Cookie{
					Name:  testHTTP.NewHeaderKey(),
					Value: testHTTP.NewHeaderValue(),
				}
				responder.SetCookie(&cookie)
				Expect(res.HeaderImpl).To(HaveKey("Set-Cookie"))
				Expect(res.HeaderImpl["Set-Cookie"]).To(ConsistOf(
					fmt.Sprintf("%s=%s", cookie.Name, cookie.Value),
				))
			})

			It("adds the expected header for multiple cookies", func() {
				cookie1 := http.Cookie{
					Name:  testHTTP.NewHeaderKey(),
					Value: testHTTP.NewHeaderValue(),
				}
				cookie2 := http.Cookie{
					Name:  testHTTP.NewHeaderKey(),
					Value: testHTTP.NewHeaderValue(),
				}
				responder.SetCookie(&cookie1)
				responder.SetCookie(&cookie2)
				Expect(res.HeaderImpl).To(HaveKey("Set-Cookie"))
				Expect(res.HeaderImpl["Set-Cookie"]).To(ConsistOf(
					fmt.Sprintf("%s=%s", cookie1.Name, cookie1.Value),
					fmt.Sprintf("%s=%s", cookie2.Name, cookie2.Value),
				))
			})
		})

		Context("Redirect", func() {
			var url string

			BeforeEach(func() {
				url = testHTTP.NewURLString()
			})

			It("responds with successful redirect", func() {
				responder.Redirect(http.StatusPermanentRedirect, url)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Location": []string{url},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{308}))
			})

			It("responds with an internal server error if there is an error with the mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				invalidMutator := request.NewHeaderMutator("", "")
				responder.Redirect(http.StatusPermanentRedirect, url, invalidMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful redirect with mutator", func() {
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				responder.Redirect(http.StatusPermanentRedirect, url, headerMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Location":        []string{url},
					headerMutator.Key: []string{headerMutator.Value}},
				))
				Expect(res.WriteHeaderInputs).To(Equal([]int{308}))
			})
		})

		Context("Empty", func() {
			It("responds with successful empty response", func() {
				responder.Empty(http.StatusOK)
				Expect(res.HeaderImpl).To(BeEmpty())
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
			})

			It("responds with an internal server error if there is an error with the mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				invalidMutator := request.NewHeaderMutator("", "")
				responder.Empty(http.StatusOK, invalidMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful empty response with mutator", func() {
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				responder.Empty(http.StatusOK, headerMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{headerMutator.Key: []string{headerMutator.Value}}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
			})
		})

		Context("Bytes", func() {
			var byts []byte

			BeforeEach(func() {
				byts = []byte(test.NewVariableString(0, 64, test.CharsetText))
			})

			It("responds with successful non-empty response", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.Bytes(http.StatusOK, byts)
				Expect(res.HeaderImpl).To(BeEmpty())
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
				Expect(res.WriteInputs).To(Equal([][]byte{byts}))
			})

			It("responds with an internal server error if there is an error with the mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				invalidMutator := request.NewHeaderMutator("", "")
				responder.Bytes(http.StatusOK, byts, invalidMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful non-empty response with mutator", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				responder.Bytes(http.StatusOK, byts, headerMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{headerMutator.Key: []string{headerMutator.Value}}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
				Expect(res.WriteInputs).To(Equal([][]byte{byts}))
			})
		})

		Context("String", func() {
			var str string

			BeforeEach(func() {
				str = test.NewVariableString(0, 64, test.CharsetText)
			})

			It("responds with successful non-empty response", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.String(http.StatusOK, str)
				Expect(res.HeaderImpl).To(BeEmpty())
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
				Expect(res.WriteInputs).To(Equal([][]byte{[]byte(str)}))
			})

			It("responds with an internal server error if there is an error with the mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				invalidMutator := request.NewHeaderMutator("", "")
				responder.String(http.StatusOK, str, invalidMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful non-empty response with mutator", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				responder.String(http.StatusOK, str, headerMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{headerMutator.Key: []string{headerMutator.Value}}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
				Expect(res.WriteInputs).To(Equal([][]byte{[]byte(str)}))
			})
		})

		Context("Reader", func() {
			var byts []byte
			var reader io.Reader

			BeforeEach(func() {
				byts = []byte(test.NewVariableString(0, 64, test.CharsetText))
				reader = bytes.NewReader(byts)
			})

			It("responds with an internal server error if the reader is missing", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.Reader(http.StatusOK, nil)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful non-empty response", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.Reader(http.StatusOK, reader)
				Expect(res.HeaderImpl).To(BeEmpty())
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
				Expect(res.WriteInputs).To(Equal([][]byte{byts}))
			})

			It("responds with an internal server error if there is an error with the mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				invalidMutator := request.NewHeaderMutator("", "")
				responder.Reader(http.StatusOK, reader, invalidMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful non-empty response with mutator", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				responder.Reader(http.StatusOK, reader, headerMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{headerMutator.Key: []string{headerMutator.Value}}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
				Expect(res.WriteInputs).To(Equal([][]byte{byts}))
			})
		})

		Context("Data", func() {
			var data *Data

			BeforeEach(func() {
				data = NewData()
			})

			It("responds with an internal server error if the data is missing", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.Data(http.StatusOK, nil)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with an internal server error if the data cannot be sanitized", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				data.SanitizeError = errors.New("sanitize error")
				responder.Data(http.StatusOK, data)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with an internal server error if the data cannot be marshalled", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				data.Value = func() {}
				responder.Data(http.StatusOK, data)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful non-empty response", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.Data(http.StatusOK, data)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
				Expect(res.WriteInputs).To(HaveLen(1))
				Expect(json.Marshal(data)).To(MatchJSON(res.WriteInputs[0]))
			})

			It("responds with an internal server error if there is an error with the mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				invalidMutator := request.NewHeaderMutator("", "")
				responder.Data(http.StatusOK, data, invalidMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful non-empty response with mutator", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				responder.Data(http.StatusOK, data, headerMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type":    []string{"application/json; charset=utf-8"},
					headerMutator.Key: []string{headerMutator.Value},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{200}))
				Expect(res.WriteInputs).To(HaveLen(1))
				Expect(json.Marshal(data)).To(MatchJSON(res.WriteInputs[0]))
			})
		})

		Context("Error", func() {
			var err *Error

			BeforeEach(func() {
				err = NewError()
			})

			It("responds with an internal server error if the error is missing", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.Error(http.StatusBadRequest, nil)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with an internal server error if the error cannot be marshalled", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				err.Value = func() {}
				responder.Error(http.StatusBadRequest, err)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful non-empty response", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.Error(http.StatusBadRequest, err)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{400}))
				Expect(res.WriteInputs).To(HaveLen(1))
				Expect(json.Marshal(err)).To(MatchJSON(res.WriteInputs[0]))
			})

			It("responds with an internal server error if there is an error with the mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				invalidMutator := request.NewHeaderMutator("", "")
				responder.Data(http.StatusBadRequest, err, invalidMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful non-empty response with mutator", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				responder.Data(http.StatusBadRequest, err, headerMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type":    []string{"application/json; charset=utf-8"},
					headerMutator.Key: []string{headerMutator.Value},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{400}))
				Expect(res.WriteInputs).To(HaveLen(1))
				Expect(json.Marshal(err)).To(MatchJSON(res.WriteInputs[0]))
			})
		})

		Context("InternalServerError", func() {
			It("responds with an internal server error if the error is missing", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.InternalServerError(nil)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with an internal server error if the error is not an internal service error", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.InternalServerError(request.ErrorUnauthenticated())
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with an internal server error if the error is an internal service error", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				responder.InternalServerError(request.ErrorInternalServerError(nil))
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with an internal server error if the error is an internal service error with mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				responder.InternalServerError(request.ErrorInternalServerError(nil), headerMutator)
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type":    []string{"application/json; charset=utf-8"},
					headerMutator.Key: []string{headerMutator.Value},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})
		})

		Context("RespondIfError", func() {
			It("returns false if error is missing", func() {
				Expect(responder.RespondIfError(nil)).To(BeFalse())
			})

			It("responds with successful non-empty response if the error is associated with a status code", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				Expect(responder.RespondIfError(request.ErrorUnauthenticated())).To(BeTrue())
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{401}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorUnauthenticated(), res.WriteInputs[0])
			})

			It("responds with successful non-empty response if the error is associated with a status code with mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				Expect(responder.RespondIfError(request.ErrorUnauthenticated(), headerMutator)).To(BeTrue())
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type":    []string{"application/json; charset=utf-8"},
					headerMutator.Key: []string{headerMutator.Value},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{401}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorUnauthenticated(), res.WriteInputs[0])
			})

			It("responds with successful non-empty response if the error is not associated with a status code", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				Expect(responder.RespondIfError(request.ErrorJSONMalformed())).To(BeTrue())
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})

			It("responds with successful non-empty response if the error is not associated with a status code with mutators", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				headerMutator := request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
				Expect(responder.RespondIfError(request.ErrorJSONMalformed(), headerMutator)).To(BeTrue())
				Expect(res.HeaderImpl).To(Equal(http.Header{
					"Content-Type":    []string{"application/json; charset=utf-8"},
					headerMutator.Key: []string{headerMutator.Value},
				}))
				Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
			})
		})
	})
})
