package hakken

import (
	"github.com/tidepool-org/go-common/clients/disc"
	"github.com/tidepool-org/go-common/jepson"
	"log"
	"net/url"
	"sync"
	"time"
)

type HakkenClient struct {
	config   HakkenClientConfig
	cooMan   coordinatorManager
	stopChan chan bool

	mut sync.Mutex
}

type HakkenClientConfig struct {
	Host              string          `json:"host"`              // Primary host to bootstrap list of coordinators from
	HeartbeatInterval jepson.Duration `json:"heartbeatInterval"` // Time elapsed between heartbeats and watch polls
	PollInterval      jepson.Duration `json:"pollInterval"`      // Time elapsed between coordinator gossip polls
	ResyncInterval    jepson.Duration `json:"resyncInterval"`    // Time elapsed between checks for new coordinators at Host
	SkipHakken        bool            `json:"skipHakken"`        // True is Hakken service is not used
}

type HakkenClientBuilder struct {
	config HakkenClientConfig
}

func NewHakkenBuilder() *HakkenClientBuilder {
	return &HakkenClientBuilder{}
}

func (b *HakkenClientBuilder) WithHost(host string) *HakkenClientBuilder {
	b.config.Host = host
	return b
}

func (b *HakkenClientBuilder) WithHeartbeatInterval(intvl time.Duration) *HakkenClientBuilder {
	b.config.HeartbeatInterval = jepson.Duration(intvl)
	return b
}

func (b *HakkenClientBuilder) WithPollInterval(intvl time.Duration) *HakkenClientBuilder {
	b.config.PollInterval = jepson.Duration(intvl)
	return b
}

func (b *HakkenClientBuilder) WithResyncInterval(intvl time.Duration) *HakkenClientBuilder {
	b.config.ResyncInterval = jepson.Duration(intvl)
	return b
}

func (b *HakkenClientBuilder) WithConfig(config *HakkenClientConfig) *HakkenClientBuilder {
	return b.WithHost(config.Host).
		WithHeartbeatInterval(time.Duration(config.HeartbeatInterval)).
		WithResyncInterval(time.Duration(config.ResyncInterval)).
		WithPollInterval(time.Duration(config.PollInterval))
}

func (b *HakkenClientBuilder) Build() *HakkenClient {
	if b.config.Host == "" {
		panic("HakkenClientBuilder requires a Host")
	}
	if b.config.HeartbeatInterval == 0 {
		b.config.HeartbeatInterval = jepson.Duration(20 * time.Second)
	}
	if b.config.PollInterval == 0 {
		b.config.PollInterval = jepson.Duration(1 * time.Minute)
	}
	if b.config.ResyncInterval == 0 {
		b.config.ResyncInterval = b.config.PollInterval * 2
	}
	return &HakkenClient{
		config: b.config,
		cooMan: coordinatorManager{
			resyncClient: coordinatorClient{Coordinator{url.URL{Scheme: "http", Host: b.config.Host}}},
			resyncTicker: time.NewTicker(time.Duration(b.config.ResyncInterval)),
			pollTicker:   time.NewTicker(time.Duration(b.config.PollInterval)),
			dropCooChan:  make(chan *coordinatorClient),
		},
		stopChan: make(chan bool),
	}
}

func (client *HakkenClient) Start() error {
	log.Println("Starting hakken")
	err := client.cooMan.start()
	if err != nil {
		return err
	}

	return nil
}

func (client *HakkenClient) Close() error {
	err := client.cooMan.Close()
	if err != nil {
		return err
	}

	close(client.stopChan)
	return nil
}

func (client *HakkenClient) Watch(service string) *disc.Watch {
	log.Printf("Creating watch for service[%s] with interval[%d]", service, client.config.HeartbeatInterval)
	slChan := make(chan *disc.Payload)
	retVal := disc.NewWatch(slChan)

	cooClient := client.cooMan.getClient()
	if cooClient != nil {
		listings, err := cooClient.getListings(service)
		if err == nil {
			done := make(chan bool)
			slChan <- disc.NewPayload(listings, done)
			<-done
		} else {
			log.Printf("Error when getting initial listings[%v]", err)
		}
	} else {
		log.Printf("No known coordinators, cannot load initial watch list for service[%s]", service)
	}

	go func() {
		timer := time.After(time.Duration(client.config.HeartbeatInterval))
		for {
			select {
			case <-client.stopChan:
				close(slChan)
			case <-timer:
				cooClient := client.cooMan.getClient()
				if cooClient != nil {
					listings, err := cooClient.getListings(service)
					if err == nil {
						done := make(chan bool)
						slChan <- disc.NewPayload(listings, done)
						<-done
					}
				}

				timer = time.After(time.Duration(client.config.HeartbeatInterval))
			}
		}
	}()

	return retVal
}

func (client *HakkenClient) Publish(sl *disc.ServiceListing) {
	log.Printf("Publishing service[%s]", sl.Service)
	go func() {
		timer := time.After(0)
		for {
			select {
			case <-client.stopChan:
				break
			case <-timer:
				for _, coo := range client.cooMan.getClients() {
					coo.listingHearbeat(sl)
				}
				timer = time.After(time.Duration(client.config.HeartbeatInterval))
			}
		}
	}()
}
