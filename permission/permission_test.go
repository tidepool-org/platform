package permission_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/permission"
)

var _ = Describe("permission", func() {
	Context("GroupIDFromUserID", func() {
		It("returns an error if the user id is missing", func() {
			groupID, err := permission.GroupIDFromUserID("", "secret")
			Expect(err).To(MatchError("permission: user id is missing"))
			Expect(groupID).To(BeEmpty())
		})

		It("returns an error if the secret is missing", func() {
			groupID, err := permission.GroupIDFromUserID("0cd1a5d68b", "")
			Expect(err).To(MatchError("permission: secret is missing"))
			Expect(groupID).To(BeEmpty())
		})

		DescribeTable("is successful for",
			func(userID string, expectedGroupID string) {
				groupID, err := permission.GroupIDFromUserID(userID, "secret")
				Expect(err).ToNot(HaveOccurred())
				Expect(groupID).To(Equal(expectedGroupID))
			},
			Entry("is example #1", "0cd1a5d68b", "NEHqFs6tA/2NRZ9oTPAHMA=="),
			Entry("is example #2", "b52201f96b", "rsWDsFcmDE2BgNfkNoiCnQ=="),
			Entry("is example #3", "46267a83eb", "cDuye1AVYPyAKvPy18+RqQ=="),
			Entry("is example #4", "982f600045", "1uO1mX4bFJ3hAC8g20l8fw=="),
			Entry("is example #5", "a06176bed7", "pMsbWdlanJldEYjkTokydA=="),
			Entry("is example #6", "d23b0a8786", "K35VY5wP6LVTpBTMUXv5OA=="),
			Entry("is example #7", "a011c16df7", "I/RdKRn3wMcaKtC/TRUIhg=="),
			Entry("is example #8", "8ea2d078f6", "AMFipBBZSHW0pP+985buzg=="),
			Entry("is example #9", "6128ef12fc", "X7DU5wxZYR9UDh780y+J9w=="),
			Entry("is example #10", "806d315a0b", "MgBbUF8XsHkj5ndZsJ0PmQ=="),
			Entry("is example #11", "aa16056cee", "iaR6v0jAWWXbDt4qS4s9HA=="),
			Entry("is example #12", "b4ba07dab4", "ARD9NlydxJZj7sJfz1UjOA=="),
			Entry("is example #13", "b4cae0bcbd", "YZGtYTIrvgSH8e7r9klFCw=="),
			Entry("is example #14", "7a1f209635", "CPzI+gdipBRYrl4ABZav4Q=="),
			Entry("is example #15", "68e70b285e", "k7kXyy3XBtoPKw9TwjLyew=="),
			Entry("is example #16", "bf33f09e3b", "HhLoSXNns8xVJh4YChWVEA=="),
			Entry("is example #17", "bb98bafa52", "4X10Q6lWGPnz2vmH7oc/6w=="),
			Entry("is example #18", "593f506db1", "ABGQBmS1eq08lnNzhMrVyg=="),
			Entry("is example #19", "480e0d76cb", "j21FL0lWNS1DU2A2dEwgMg=="),
			Entry("is example #20", "970d79a164", "3CyaEVxSX0HgvBCwEHiSBg=="),
		)
	})
})
