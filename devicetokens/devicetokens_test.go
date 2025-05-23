package devicetokens

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
)

const mockUserID1 = "008c7f79-6545-4466-95fb-34e3ba728d38"

func TestSuite(t *testing.T) {
	test.Test(t)
}

var _ = Describe("DeviceToken", func() {
	It("parses", func() {
		buf := buff(`{"apple":{"token":"c29tZXRoaW5n","environment":"sandbox"}}`)
		token := &DeviceToken{}
		err := request.DecodeObject(context.Background(), nil, buf, token)
		Expect(err).ToNot(HaveOccurred())
	})

	It("validates environment", func() {
		token := &DeviceToken{}
		bad := buff(`{"apple":{"token":"c29tZXRoaW5n","environment":"bad"}}`)
		err := request.DecodeObject(context.Background(), nil, bad, token)
		Expect(err).To(MatchError("value \"bad\" is not one of [\"production\", \"sandbox\"]"))

		prod := buff(`{"apple":{"token":"c29tZXRoaW5n","environment":"production"}}`)
		err = request.DecodeObject(context.Background(), nil, prod, token)
		Expect(err).ToNot(HaveOccurred())

		sbox := buff(`{"apple":{"token":"c29tZXRoaW5n","environment":"sandbox"}}`)
		err = request.DecodeObject(context.Background(), nil, sbox, token)
		Expect(err).ToNot(HaveOccurred())
	})

	It("validates token", func() {
		token := &DeviceToken{}
		buf := buff(`{"apple":{"token":"","environment":"sandbox"}}`)
		err := request.DecodeObject(context.Background(), nil, buf, token)
		Expect(err).To(MatchError("value is empty"))

		buf = buff(`{"apple":{"token":"not-base64","environment":"sandbox"}}`)
		err = request.DecodeObject(context.Background(), nil, buf, token)
		Expect(err).To(MatchError("json is malformed"))

		testToken := buildTokenLongerThan(MaxTokenLen)
		buf = buff(`{"apple":{"token":"%s","environment":"sandbox"}}`, testToken)
		err = request.DecodeObject(context.Background(), nil, buf, token)
		Expect(err).To(MatchError(ContainSubstring("is not less than or equal to")))
	})

	It("apple must exist (there's no other supported device yet)", func() {
		token := &DeviceToken{}
		buf := buff(`{}`)
		err := request.DecodeObject(context.Background(), nil, buf, token)
		Expect(err).To(MatchError(ContainSubstring("value does not exist")))
	})

	Describe("NewDocument", func() {
		It("generates a TokenID", func() {
			token := DeviceToken{
				Apple: &AppleDeviceToken{Environment: "sandbox", Token: []byte("blah")},
			}
			doc := NewDocument(mockUserID1, token)

			Expect(doc.TokenKey).To(HaveLen(64))
			Expect(doc.TokenKey).To(MatchRegexp("[a-fA-F0-9]{64}"))
		})
	})

	Describe("key", func() {
		It("produces a hash", func() {
			token := &DeviceToken{}
			buf := buff(`{"apple":{"token":"c29tZXRoaW5n","environment":"sandbox"}}`)
			err := request.DecodeObject(context.Background(), nil, buf, token)
			Expect(err).ToNot(HaveOccurred())
			key := token.key()
			Expect(key).To(HaveLen(64))
			Expect(key).To(MatchRegexp("[a-fA-F0-9]{64}"))
		})
	})
})

// buff is a helper for generating a JSON []byte representation.
func buff(format string, args ...interface{}) *bytes.Buffer {
	return bytes.NewBufferString(fmt.Sprintf(format, args...))
}

// buildTokenLongerThan builds a token that's just TOO long.
func buildTokenLongerThan(limit int) string {
	var tooLong = []byte{}
	for i := 0; i < limit+1; i++ {
		tooLong = append(tooLong, 'a')
	}
	return base64.StdEncoding.EncodeToString(tooLong)
}
