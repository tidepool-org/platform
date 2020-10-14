package disc

import (
	"log"
	"net/url"
)

type HostGetter interface {
	HostGet() []url.URL
}

type HostGetterFunc func() []url.URL

func (h HostGetterFunc) HostGet() []url.URL {
	return h()
}

type StaticHostGetter struct {
	Hosts []url.URL
}

func NewStaticHostGetter(retVal url.URL) *StaticHostGetter {
	return &StaticHostGetter{Hosts: []url.URL{retVal}}
}

func NewStaticHostGetterFromString(urlString string) *StaticHostGetter {
	theUrl, err := url.Parse(urlString)
	if err != nil {
		log.Printf("Unable to parse urlString[%s]", urlString)
		return nil
	}
	return NewStaticHostGetter(*theUrl)
}

func (h *StaticHostGetter) HostGet() []url.URL {
	return h.Hosts
}
