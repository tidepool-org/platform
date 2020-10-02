package disc

import (
	"log"
	"math/rand"
	"net/url"
	"sync"
	"time"
)

var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type Watch struct {
	incoming chan *Payload
	listings []ServiceListing

	mut sync.RWMutex
}

type Payload struct {
	listings []ServiceListing
	done     chan bool
}

func NewPayload(listings []ServiceListing, done chan bool) *Payload {
	return &Payload{listings: listings, done: done}
}

func NewWatch(theChan chan *Payload) *Watch {
	retVal := &Watch{incoming: theChan}
	retVal.start()
	return retVal
}

func (g *Watch) ServiceListingsGet() []ServiceListing {
	g.mut.RLock()
	defer g.mut.RUnlock()
	return g.listings
}

func (g *Watch) start() {
	go func() {
		more := true
		for more {
			var payload *Payload
			payload, more = <-g.incoming
			theList := payload.listings
			addedItems := make([]ServiceListing, len(theList))
			copy(addedItems, theList)
			var removedItems []ServiceListing

			g.mut.Lock()
			for _, listing := range g.listings {
				found := false
				for i, newListing := range addedItems {
					if newListing.Equals(listing) {
						addedItems = append(addedItems[0:i], addedItems[i+1:]...)
						found = true
						break
					}
				}
				if !found {
					removedItems = append(removedItems, listing)
				}
			}
			g.listings = theList
			g.mut.Unlock()

			for _, listing := range removedItems {
				log.Printf("Removing listing[%+v]", listing)
			}
			for _, listing := range addedItems {
				log.Printf("Adding listing[%+v]", listing)
			}
			close(payload.done)
		}
	}()
}

func (g *Watch) Random() HostGetter {
	return HostGetterFunc(func() []url.URL {
		listings := g.ServiceListingsGet()
		if len(listings) == 0 {
			return nil
		}

		return []url.URL{listings[random.Intn(len(listings))].URL}
	})
}
