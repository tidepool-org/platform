package client

import "net/http"

type RoundTripper struct {
	roundTripper http.RoundTripper
}

func NewRoundTripper(roundTripper http.RoundTripper) *RoundTripper {
	return &RoundTripper{
		roundTripper: roundTripper,
	}
}

func (p *RoundTripper) ResolvedRoundTripper() http.RoundTripper {
	roundTripper := p.roundTripper
	if roundTripper == nil {
		roundTripper = http.DefaultClient.Transport
		if roundTripper == nil {
			roundTripper = http.DefaultTransport
		}
	}
	return roundTripper
}

func (p *RoundTripper) WithRoundTripper(roundTripper http.RoundTripper) {
	p.roundTripper = roundTripper
}

func (p *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return p.ResolvedRoundTripper().RoundTrip(req)
}
