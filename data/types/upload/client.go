package upload

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/structure"
)

type Client struct {
	Name    *string            `json:"name,omitempty" bson:"name,omitempty"`
	Version *string            `json:"version,omitempty" bson:"version,omitempty"`
	Private *metadata.Metadata `json:"private,omitempty" bson:"private,omitempty"`
}

func ParseClient(parser structure.ObjectParser) *Client {
	if !parser.Exists() {
		return nil
	}
	datum := NewClient()
	parser.Parse(datum)
	return datum
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Parse(parser structure.ObjectParser) {
	c.Name = parser.String("name")
	c.Version = parser.String("version")
	c.Private = metadata.ParseMetadata(parser.WithReferenceObjectParser("private"))
}

func (c *Client) Validate(validator structure.Validator) {
	validator.String("name", c.Name).Exists().Using(net.ReverseDomainValidator)
	validator.String("version", c.Version).Exists().Using(net.SemanticVersionValidator)
	if c.Private != nil {
		c.Private.Validate(validator.WithReference("private"))
	}
}

func (c *Client) Normalize(normalizer data.Normalizer) {}
