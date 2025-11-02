package config

import "os"

type Config struct {
	GroqAPIKey string
	Port       string
}

func Load() *Config {
	groqAPIKey := os.Getenv("GROQ_API_KEY")
	if groqAPIKey == "" {
		panic("GROQ_API_KEY environment variable not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		GroqAPIKey: groqAPIKey,
		Port:       port,
	}
}
