package metrics

import (
	"errors"
	"fmt"

	"github.com/nicolasbonnici/gorest/database"
)

type Config struct {
	Database           database.Database
	AllowedTypes       []string `json:"allowed_types" yaml:"allowed_types"`
	MaxKeyLength       int      `json:"max_key_length" yaml:"max_key_length"`
	OnlyPositiveValues bool     `json:"only_positive_values" yaml:"only_positive_values"`
	PaginationLimit    int      `json:"pagination_limit" yaml:"pagination_limit"`
	MaxPaginationLimit int      `json:"max_pagination_limit" yaml:"max_pagination_limit"`
}

func DefaultConfig() Config {
	return Config{
		AllowedTypes:       []string{"post"},
		MaxKeyLength:       255,
		OnlyPositiveValues: false,
		PaginationLimit:    50,
		MaxPaginationLimit: 200,
	}
}

func (c *Config) Validate() error {
	if len(c.AllowedTypes) == 0 {
		return errors.New("allowed_types cannot be empty")
	}

	seen := make(map[string]bool)
	for _, resourceType := range c.AllowedTypes {
		if resourceType == "" {
			return errors.New("allowed_types cannot contain empty strings")
		}
		if seen[resourceType] {
			return fmt.Errorf("duplicate type in allowed_types: %s", resourceType)
		}
		seen[resourceType] = true
	}

	if c.MaxKeyLength < 1 || c.MaxKeyLength > 255 {
		return errors.New("max_key_length must be between 1 and 255")
	}

	if c.PaginationLimit < 1 || c.PaginationLimit > c.MaxPaginationLimit {
		return errors.New("pagination_limit must be between 1 and max_pagination_limit")
	}

	if c.MaxPaginationLimit < 1 || c.MaxPaginationLimit > 1000 {
		return errors.New("max_pagination_limit must be between 1 and 1000")
	}

	return nil
}

func (c *Config) IsAllowedType(resourceType string) bool {
	for _, allowed := range c.AllowedTypes {
		if allowed == resourceType {
			return true
		}
	}
	return false
}
