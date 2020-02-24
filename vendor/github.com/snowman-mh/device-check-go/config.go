package devicecheck

// Environment specifies base URL for DeviceCheck API
type Environment int

const (
	// Development Environment
	Development Environment = iota + 1
	// Production Environment
	Production
)

// Config provides configuration for DeviceCheck API
type Config struct {
	env    Environment
	issuer string
	keyID  string
}

// NewConfig returns a new configuration instance
func NewConfig(issuer, keyID string, env Environment) Config {
	return Config{
		env:    env,
		issuer: issuer,
		keyID:  keyID,
	}
}
