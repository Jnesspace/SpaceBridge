// Package config handles configuration management for SpaceBridge.
package config

import (
	"errors"
	"os"
)

// AccountConfig holds the configuration for a single Spacelift account.
type AccountConfig struct {
	URL       string
	KeyID     string
	SecretKey string
}

// Config holds the complete SpaceBridge configuration.
type Config struct {
	Source      AccountConfig
	Destination AccountConfig
}

// Validate checks if the configuration has all required fields.
func (c *AccountConfig) Validate() error {
	if c.URL == "" {
		return errors.New("spacelift URL is required")
	}
	if c.KeyID == "" {
		return errors.New("spacelift API key ID is required")
	}
	if c.SecretKey == "" {
		return errors.New("spacelift API secret key is required")
	}
	return nil
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Source: AccountConfig{
			URL:       os.Getenv("SOURCE_SPACELIFT_URL"),
			KeyID:     os.Getenv("SOURCE_SPACELIFT_KEY_ID"),
			SecretKey: os.Getenv("SOURCE_SPACELIFT_SECRET_KEY"),
		},
		Destination: AccountConfig{
			URL:       os.Getenv("DESTINATION_SPACELIFT_URL"),
			KeyID:     os.Getenv("DESTINATION_SPACELIFT_KEY_ID"),
			SecretKey: os.Getenv("DESTINATION_SPACELIFT_SECRET_KEY"),
		},
	}

	return cfg, nil
}

// ValidateSource validates that source account configuration is complete.
func (c *Config) ValidateSource() error {
	return c.Source.Validate()
}

// ValidateDestination validates that destination account configuration is complete.
func (c *Config) ValidateDestination() error {
	return c.Destination.Validate()
}

// HasDestination returns true if destination configuration is provided.
func (c *Config) HasDestination() bool {
	return c.Destination.URL != "" && c.Destination.KeyID != "" && c.Destination.SecretKey != ""
}
