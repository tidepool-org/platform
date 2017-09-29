package upload

import "github.com/tidepool-org/platform/data"

type Client struct {
	Name    *string `json:"name,omitempty" bson:"name,omitempty"`
	Version *string `json:"version,omitempty" bson:"version,omitempty"`
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Parse(parser data.ObjectParser) {
	c.Name = parser.ParseString("name")
	c.Version = parser.ParseString("version")
}

func (c *Client) Validate(validator data.Validator) {
	validator.ValidateString("name", c.Name).Exists()       // TODO: Additional validation
	validator.ValidateString("version", c.Version).Exists() // TODO: Additional validation
}

func (c *Client) Normalize(normalizer data.Normalizer) {
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
