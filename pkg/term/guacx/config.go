package guacx

type Config struct {
	ConnectionID string
	Protocol     string
	Parameters   map[string]string
}

func NewConfig() (config *Config) {
	config = &Config{}
	config.Parameters = make(map[string]string)
	return config
}

func (c *Config) SetParameter(name, value string) {
	c.Parameters[name] = value
}

func (c *Config) GetParameter(name string) string {
	return c.Parameters[name]
}
