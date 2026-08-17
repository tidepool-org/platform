package work_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	oauthWorkTest "github.com/tidepool-org/platform/oauth/work/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("token", func() {
	Context("TokenMetadata", func() {
		Context("MetadataKeyOAuthToken", func() {
			It("returns expected value", func() {
				Expect(oauthWork.MetadataKeyOAuthToken).To(Equal("oauthToken"))
			})
		})

		Context("TokenMetadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *oauthWork.TokenMetadata)) {
					datum := oauthWorkTest.RandomTokenMetadata(test.AllowOptionals())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, oauthWorkTest.NewObjectFromTokenMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, oauthWorkTest.NewObjectFromTokenMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *oauthWork.TokenMetadata) {},
				),
				Entry("empty",
					func(datum *oauthWork.TokenMetadata) {
						*datum = oauthWork.TokenMetadata{}
					},
				),
				Entry("all",
					func(datum *oauthWork.TokenMetadata) {
						datum.OAuthToken = authTest.RandomToken()
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *oauthWork.TokenMetadata), expectedErrors ...error) {
						expectedDatum := oauthWorkTest.RandomTokenMetadata(test.AllowOptionals())
						object := oauthWorkTest.NewObjectFromTokenMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &oauthWork.TokenMetadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *oauthWork.TokenMetadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *oauthWork.TokenMetadata) {
							clear(object)
							*expectedDatum = oauthWork.TokenMetadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *oauthWork.TokenMetadata) {
							object["oauthToken"] = true
							expectedDatum.OAuthToken = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/oauthToken"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *oauthWork.TokenMetadata), expectedErrors ...error) {
						datum := oauthWorkTest.RandomTokenMetadata(test.AllowOptionals())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *oauthWork.TokenMetadata) {},
					),
					Entry("oauth token missing",
						func(datum *oauthWork.TokenMetadata) {
							datum.OAuthToken = nil
						},
					),
					Entry("oauth token invalid",
						func(datum *oauthWork.TokenMetadata) {
							datum.OAuthToken = authTest.RandomToken()
							datum.OAuthToken.AccessToken = ""
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/oauthToken/accessToken"),
					),
					Entry("oauth token valid",
						func(datum *oauthWork.TokenMetadata) {
							datum.OAuthToken = authTest.RandomToken()
						},
					),
					Entry("multiple errors",
						func(datum *oauthWork.TokenMetadata) {
							datum.OAuthToken = authTest.RandomToken()
							datum.OAuthToken.AccessToken = ""
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/oauthToken/accessToken"),
					),
				)
			})
		})
	})
})
