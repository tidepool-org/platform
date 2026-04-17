package user_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/user"
)

var _ = Describe("User", func() {
	Context("LegacySeagullDocument", func() {
		Context("AddProfileToSeagullValue", func() {
			It("Preserves non profile seagull fields such as settings, etc", func() {
				seagullValueBefore := `{
				"profile": {"fullName": "something"},
				"preferences": { "clickedUploaderBannerTime": "2023-01-10T10:11:12-08:00" },
				"settings": { "bgTarget": { "high": 160, "low":  60 }, "units": { "bg": "mg/dL" } }
			}`
				addedProfile := &user.LegacyUserProfile{
					FullName: "Some Name",
					Patient: &user.LegacyPatientProfile{
						Birthday:      "2000-03-04",
						DiagnosisDate: "2001-03-05",
						About:         "About me",
					},
					MigrationStatus: user.MigrationCompleted,
				}
				expectedNewSeagullValue := `{
					"profile": {"fullName": "Some Name", "patient": { "birthday": "2000-03-04", "diagnosisDate": "2001-03-05", "about": "About me"}},
					"preferences": { "clickedUploaderBannerTime": "2023-01-10T10:11:12-08:00" },
					"settings": { "bgTarget": { "high": 160, "low":  60 }, "units": { "bg": "mg/dL" } }}`

				newValue, err := user.AddProfileToSeagullValue(seagullValueBefore, addedProfile)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(newValue).To(MatchJSON(expectedNewSeagullValue))
			})
		})
	})

	Context("Profile", func() {
		DescribeTable("ToLegacyProfile",
			func(profile *user.UserProfile, legacyProfile *user.LegacyUserProfile, roles []string) {
				Expect(profile.ToLegacyProfile(roles)).To(BeComparableTo(legacyProfile))
			},
			Entry("Regular patient", &user.UserProfile{
				FullName:       "Bob",
				Birthday:       "2000-02-03",
				About:          "About me",
				MRN:            "1112222",
				TargetDevices:  []string{"SomeDevice900"},
				TargetTimezone: "UTC",
			},
				&user.LegacyUserProfile{
					FullName: "Bob",
					Patient: &user.LegacyPatientProfile{
						Birthday:       "2000-02-03",
						About:          "About me",
						MRN:            "1112222",
						TargetDevices:  []string{"SomeDevice900"},
						TargetTimezone: "UTC",
					},
					MigrationStatus: user.MigrationCompleted,
				},
				[]string{user.RolePatient},
			),
			Entry("Fake child", &user.UserProfile{
				FullName:      "Child Name",
				Birthday:      "2000-02-03",
				DiagnosisDate: "2001-02-03",
				About:         "About me",
				Custodian: &user.Custodian{
					FullName: "Parent Name",
				},
			},
				&user.LegacyUserProfile{
					FullName: "Parent Name",
					Patient: &user.LegacyPatientProfile{
						FullName:      pointer.FromString("Child Name"),
						Birthday:      "2000-02-03",
						DiagnosisDate: "2001-02-03",
						About:         "About me",
						IsOtherPerson: true,
					},
					MigrationStatus: user.MigrationCompleted,
				},
				[]string{user.RolePatient},
			),
			Entry("Clinic", &user.UserProfile{
				FullName: "Clinician Name",
				Clinic: &user.ClinicProfile{
					Name:      pointer.FromString("Clinic Name"),
					Role:      pointer.FromString("Some Role"),
					Telephone: pointer.FromString("123-123-3456"),
					NPI:       pointer.FromString("1234567890"),
				},
			},
				&user.LegacyUserProfile{
					FullName: "Clinician Name",
					Clinic: &user.ClinicProfile{
						Name:      pointer.FromString("Clinic Name"),
						Role:      pointer.FromString("Some Role"),
						Telephone: pointer.FromString("123-123-3456"),
						NPI:       pointer.FromString("1234567890"),
					},
					MigrationStatus: user.MigrationCompleted,
				},
				[]string{user.RoleClinician},
			),
		)
	})
})
