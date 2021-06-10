package devicecheck

// Environment specifies DeviceCheck API environment.
type Environment int

const (
	// Development specifies Apple's development environment.
	Development Environment = iota + 1
	// Production specifies Apple's production environment.
	Production
)

// Config provides configuration for DeviceCheck API.
type Config struct {
	env    Environment
	issuer string
	keyID  string
}

// NewConfig returns a new configuration.
func NewConfig(issuer, keyID string, env Environment) Config {
	return Config{
		env:    env,
		issuer: issuer,
		keyID:  keyID,
	}
}
