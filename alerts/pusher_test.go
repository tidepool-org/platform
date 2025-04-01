package alerts

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("APNSPusher", func() {
	Describe("NewAPNSPusherFromEnv", func() {
		It("succeeds", func() {
			configureEnvconfig()
			pusher, err := NewPusher()
			Expect(err).To(Succeed())
			Expect(pusher).ToNot(Equal(nil))
		})
	})
})

var _ = Describe("LoadAPNSPusherConfigFromEnv", func() {
	BeforeEach(func() {
		configureEnvconfig()
	})

	It("errors if key data is empty or blank", func() {
		GinkgoT().Setenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_SIGNING_KEY", "")
		_, err := NewPusher()
		Expect(err).To(MatchError(ContainSubstring("APNs signing key is blank")))

		os.Unsetenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_SIGNING_KEY")
		_, err = NewPusher()
		Expect(err).To(MatchError(ContainSubstring("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_SIGNING_KEY missing value")))
	})

	It("errors if key data is invalid", func() {
		GinkgoT().Setenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_SIGNING_KEY", "invalid")
		_, err := NewPusher()
		Expect(err).To(MatchError(ContainSubstring("AuthKey must be a valid .p8 PEM file")))
	})

	It("errors if bundleID is blank", func() {
		GinkgoT().Setenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_BUNDLE_ID", "")
		_, err := NewPusher()
		Expect(err).To(MatchError(ContainSubstring("bundleID is blank")))
	})

	It("errors if teamID is blank", func() {
		GinkgoT().Setenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_TEAM_ID", "")
		_, err := NewPusher()
		Expect(err).To(MatchError(ContainSubstring("teamID is blank")))
	})

	It("errors if keyID is blank", func() {
		GinkgoT().Setenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_KEY_ID", "")
		_, err := NewPusher()
		Expect(err).To(MatchError(ContainSubstring("keyID is blank")))
	})
})

func configureEnvconfig() {
	GinkgoT().Setenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_SIGNING_KEY", string(validTestKey))
	GinkgoT().Setenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_KEY_ID", "key")
	GinkgoT().Setenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_TEAM_ID", "team")
	GinkgoT().Setenv("TIDEPOOL_CARE_PARTNER_ALERTS_APNS_BUNDLE_ID", "bundle")
}

// validTestKey is a random private key for testing
var validTestKey = []byte(`-----BEGIN PRIVATE KEY-----
MIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDDNrXT9ZRWPUAAg38Qi
Z553y7sGqOgMxUCG36eCIcRCy1QiTJBgGDxIhWvkE8Sx4N6hZANiAATrsRyRXLa0
Tgczq8tmFomMP212HdkPF3gFEl/CkqGHUodR2EdZBW1zVcmuLjIN4zvqVVXMJm/U
eHZz9xAZ95y3irAfkMuOD/Bw88UYvhKnipOHBeS8BwqyfFQ+NRB6xYU=
-----END PRIVATE KEY-----
`)
