package main

import (
	"fmt"
	"os"
)

// Reads in the app password from the environment variable PUNKBOT_PASSWORD
func (c *Config) GetSecretFromEnv() error {
	c.password = os.Getenv("PUNKBOT_PASSWORD")
	if c.password == "" {
		return fmt.Errorf("no password set, PUNKBOT_PASSWORD is probably not set")
	}
	return nil
}

func (c *Config) GetSecret() string {
	return c.password
}
