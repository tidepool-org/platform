package queue

import (
	"fmt"
	"math/rand/v2"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LRU Queue", func() {
	Context("Store", func() {
		It("has a limited size", func() {
			maxSize := 1 + rand.IntN(1000)
			base := time.Now()
			c := NewMRUCache(maxSize)
			for i := range maxSize * 2 {
				c.Store(fmt.Sprintf("%d", i), base.Add(time.Duration(i)))
			}
			Expect(len(c.cache)).To(Equal(maxSize))
			Expect(len(c.sorted)).To(Equal(maxSize))
		})

		It("deletes the oldest entry first", func() {
			maxSize := 1 + rand.IntN(1000)
			base := time.Now()
			c := NewMRUCache(maxSize)
			for i := range maxSize + 1 {
				c.Store(fmt.Sprintf("%d", i), base.Add(time.Duration(i)))
			}
			Expect(len(c.cache)).To(Equal(maxSize))
			Expect(len(c.sorted)).To(Equal(maxSize))
			Expect(c.Get("0")).To(Equal(time.Time{}))
			Expect(c.Get("1")).To(Equal(base.Add(1)))
		})

		It("has its last used time refreshed when stored again", func() {
			maxSize := 1 + rand.IntN(1000)
			base := time.Now()
			c := NewMRUCache(maxSize)
			for i := range maxSize {
				c.Store(fmt.Sprintf("%d", i), base.Add(time.Duration(i)))
			}
			c.Store("0", base)                        // refresh "0" in the LRU
			c.Store("too many", base)                 // bump "1" out of the LRU
			Expect(c.Get("0")).To(Equal(base))        // "0" is still there
			Expect(c.Get("1")).To(Equal(time.Time{})) // "1" is gone
			Expect(c.Get("2")).To(Equal(base.Add(2))) // "2" is (still) there
		})
	})

	Context("Get", func() {
		It("refreshes the value's last used order", func() {
			maxSize := 1 + rand.IntN(1000)
			base := time.Now()
			c := NewMRUCache(maxSize)
			for i := range maxSize {
				c.Store(fmt.Sprintf("%d", i), base.Add(time.Duration(i)))
			}
			c.Get("0")
			c.Store("too many", base)
			Expect(c.Get("0")).To(Equal(base))
			Expect(c.Get("1")).To(Equal(time.Time{}))
		})
	})
})
