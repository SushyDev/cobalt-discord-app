package config

// Config holds application configuration parameters
type Config struct {
	DiscordToken string
	CobaltURL    string
	CobaltAuthKey string
}

// NewConfig creates a new configuration
func NewConfig(discordToken, cobaltURL, cobaltAuthKey string) *Config {
	return &Config{
		DiscordToken: discordToken,
		CobaltURL:    cobaltURL,
		CobaltAuthKey: cobaltAuthKey,
	}
}