package upload

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/structure"
)

type Client struct {
	Name    *string    `json:"name,omitempty" bson:"name,omitempty"`
	Version *string    `json:"version,omitempty" bson:"version,omitempty"`
	Private *data.Blob `json:"private,omitempty" bson:"private,omitempty"`
}

func ParseClient(parser data.ObjectParser) *Client {
	if parser.Object() == nil {
		return nil
	}
	client := NewClient()
	client.Parse(parser)
	parser.ProcessNotParsed()
	return client
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Parse(parser data.ObjectParser) {
	c.Name = parser.ParseString("name")
	c.Version = parser.ParseString("version")
	c.Private = data.ParseBlob(parser.NewChildObjectParser("private"))
}

func (c *Client) Validate(validator structure.Validator) {
	validator.String("name", c.Name).Exists().Using(net.ReverseDomainValidator)
	validator.String("version", c.Version).Exists().Using(net.SemanticVersionValidator)
	if c.Private != nil {
		c.Private.Validate(validator.WithReference("private"))
	}
}

func (c *Client) Normalize(normalizer data.Normalizer) {}
