package udi_test

import (
	"encoding/json"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/cluster"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/udi"
)

const commitHash = "868bd7b1c4f0e3a5d2b9c8e7f6a5b4c3d2e1f0a9"

var _ = Describe("UDI", func() {
	It("has a device identifier that is a GTIN-14 with a valid check digit", func() {
		Expect(udi.DeviceIdentifier).To(MatchRegexp(`^[0-9]{14}$`))
		sum := 0
		for index, digit := range udi.DeviceIdentifier[:13] {
			weight := 1
			if index%2 == 0 {
				weight = 3
			}
			sum += int(digit-'0') * weight
		}
		Expect(int(udi.DeviceIdentifier[13] - '0')).To(Equal((10 - sum%10) % 10))
	})

	It("limits the frontend version to what fits the software version", func() {
		Expect(udi.FrontendVersionLengthMaximum).To(Equal(12))
	})

	Context("New", func() {
		release := &cluster.Release{Revision: commitHash, Time: time.Date(2026, 11, 3, 18, 42, 7, 0, time.UTC)}

		It("returns the approved example", func() {
			Expect(udi.New(release, "1.103.2")).To(Equal(&udi.UDI{
				ProductName:         "Tidepool Data Platform",
				ManufacturerName:    "Tidepool Project",
				ManufacturerAddress: "555 Bryant St., #429, Palo Alto, CA 94301",
				ReleaseDate:         "2026-11-03",
				ReleaseIdentifier:   "868bd7b",
				SoftwareVersion:     "1.103.2-868bd7b",
				DeviceIdentifier:    "00884400418012",
				UDI:                 "(01)00884400418012(11)261103(8012)1.103.2-868bd7b",
			}))
		})

		It("keeps the leading zeros of the device identifier as a JSON string", func() {
			result, err := udi.New(release, "1.103.2")
			Expect(err).ToNot(HaveOccurred())
			bites, err := json.Marshal(result)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bites)).To(ContainSubstring(`"deviceIdentifier":"00884400418012"`))
			Expect(string(bites)).To(ContainSubstring(`"udi":"(01)00884400418012(11)`))
		})

		DescribeTable("formats the release date as YYMMDD and YYYY-MM-DD in UTC",
			func(releaseTime time.Time, expectedDate string, expectedReleaseDate string) {
				result, err := udi.New(&cluster.Release{Revision: commitHash, Time: releaseTime}, "1.103.2")
				Expect(err).ToNot(HaveOccurred())
				Expect(result.UDI).To(Equal("(01)00884400418012(11)" + expectedDate + "(8012)1.103.2-868bd7b"))
				Expect(result.ReleaseDate).To(Equal(expectedReleaseDate))
			},
			Entry("single digit month and day", time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC), "270105", "2027-01-05"),
			Entry("last second of the day", time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC), "261231", "2026-12-31"),
			Entry("Pacific evening is the next day in UTC", time.Date(2026, 11, 3, 17, 0, 0, 0, time.FixedZone("PST", -8*60*60)), "261104", "2026-11-04"),
		)

		It("allows a software version of 20 characters", func() {
			result, err := udi.New(release, "10.100.10000")
			Expect(err).ToNot(HaveOccurred())
			Expect(result.SoftwareVersion).To(HaveLen(20))
			Expect(result.UDI).To(HaveSuffix("(8012)10.100.10000-868bd7b"))
		})

		DescribeTable("returns an error if the identification data is missing or invalid",
			func(release *cluster.Release, frontendVersion string, expectedError string) {
				result, err := udi.New(release, frontendVersion)
				Expect(err).To(MatchError(expectedError))
				Expect(result).To(BeNil())
			},
			Entry("release missing", nil, "1.103.2", "release is missing"),
			Entry("release time missing", &cluster.Release{Revision: commitHash}, "1.103.2", "release time is missing"),
			Entry("release revision too short", &cluster.Release{Revision: "868bd7", Time: time.Now()}, "1.103.2", `release revision "868bd7" is invalid`),
			Entry("frontend version missing", &cluster.Release{Revision: commitHash, Time: time.Now()}, "", "frontend version is missing"),
			Entry("software version longer than 20 characters", &cluster.Release{Revision: commitHash, Time: time.Now()}, "10.100.100000", `software version "10.100.100000-868bd7b" is longer than 20 characters`),
		)
	})

	Context("Request", func() {
		expression := udi.FrontendVersionExpression

		DescribeTable("validates the frontend version",
			func(frontendVersion *string, expectedErrors ...error) {
				request := udi.NewRequest()
				request.FrontendVersion = frontendVersion
				errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(request), expectedErrors...)
			},
			Entry("valid", pointer.FromString("1.103.2")),
			Entry("valid with a prerelease and build", pointer.FromString("1.103.2-rc.1")),
			Entry("valid at 12 characters", pointer.FromString(strings.Repeat("1", 12))),
			Entry("missing", nil,
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/frontendVersion"),
			),
			Entry("empty", pointer.FromString(""),
				errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, 1, 12), "/frontendVersion"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotMatches("", expression), "/frontendVersion"),
			),
			Entry("longer than 12 characters", pointer.FromString(strings.Repeat("1", 13)),
				errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(13, 1, 12), "/frontendVersion"),
			),
			Entry("GS1 application identifier", pointer.FromString("1.0(01)2"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotMatches("1.0(01)2", expression), "/frontendVersion"),
			),
			Entry("space", pointer.FromString("1.103 2"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotMatches("1.103 2", expression), "/frontendVersion"),
			),
		)
	})
})
