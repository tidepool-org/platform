package clients

import (
	"github.com/tidepool-org/go-common/clients/disc"
	"github.com/tidepool-org/go-common/clients/hakken"
	"github.com/tidepool-org/go-common/clients/highwater"
	"github.com/tidepool-org/go-common/clients/shoreline"
	"log"
	"net/url"
)

type HostGetterConfig interface{}

func ToHostGetter(name string, c *HostGetterConfig, discovery disc.Discovery) disc.HostGetter {
	switch c := (*c).(type) {
	case string:
		return discovery.Watch(c).Random()
	case map[string]interface{}:
		theType := c["type"].(string)
		switch theType {
		case "static":
			hostStrings := c["hosts"].([]interface{})
			hosts := make([]url.URL, len(hostStrings))
			for i, v := range hostStrings {
				host, err := url.Parse(v.(string))
				if err != nil {
					panic(err.Error())
				}
				hosts[i] = *host
			}

			log.Printf("service[%s] with static watch for hosts[%v]", name, hostStrings)
			return &disc.StaticHostGetter{Hosts: hosts}
		case "random":
			return discovery.Watch(c["service"].(string)).Random()
		}
	default:
		log.Panicf("Unexpected type for HostGetterConfig[%T]", c)
	}

	panic("Appease the compiler, code should never get here")
}

type GatekeeperConfig struct {
	HostGetter HostGetterConfig `json:"serviceSpec"`
}

func (gc *GatekeeperConfig) ToHostGetter(discovery disc.Discovery) disc.HostGetter {
	return ToHostGetter("gatekeeper", &gc.HostGetter, discovery)
}

type SeagullConfig struct {
	HostGetter HostGetterConfig `json:"serviceSpec"`
}

func (sc *SeagullConfig) ToHostGetter(discovery disc.Discovery) disc.HostGetter {
	return ToHostGetter("seagull", &sc.HostGetter, discovery)
}

type ShorelineConfig struct {
	shoreline.ShorelineClientConfig
	HostGetter HostGetterConfig `json:"serviceSpec"`
}

func (uac *ShorelineConfig) ToHostGetter(discovery disc.Discovery) disc.HostGetter {
	return ToHostGetter("user-api", &uac.HostGetter, discovery)
}

type HighwaterConfig struct {
	highwater.HighwaterClientConfig
	HostGetter HostGetterConfig `json:"serviceSpec"`
}

func (hc *HighwaterConfig) ToHostGetter(discovery disc.Discovery) disc.HostGetter {
	return ToHostGetter("highwater", &hc.HostGetter, discovery)
}

type Config struct {
	HakkenConfig     hakken.HakkenClientConfig `json:"hakken"`
	GatekeeperConfig GatekeeperConfig          `json:"gatekeeper"`
	SeagullConfig    SeagullConfig             `json:"seagull"`
	ShorelineConfig  ShorelineConfig           `json:"shoreline"`
	HighwaterConfig  HighwaterConfig           `json:"highwater"`
}
