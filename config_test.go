package metrics

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				AllowedTypes:       []string{"post", "user"},
				MaxKeyLength:       255,
				OnlyPositiveValues: false,
				PaginationLimit:    50,
				MaxPaginationLimit: 200,
			},
			wantErr: false,
		},
		{
			name: "empty allowed types",
			config: Config{
				AllowedTypes: []string{},
			},
			wantErr: true,
			errMsg:  "allowed_types cannot be empty",
		},
		{
			name: "allowed types with empty string",
			config: Config{
				AllowedTypes: []string{"post", ""},
			},
			wantErr: true,
			errMsg:  "allowed_types cannot contain empty strings",
		},
		{
			name: "duplicate type names",
			config: Config{
				AllowedTypes: []string{"post", "user", "post"},
			},
			wantErr: true,
			errMsg:  "duplicate type in allowed_types: post",
		},
		{
			name: "max key length too small",
			config: Config{
				AllowedTypes:       []string{"post"},
				MaxKeyLength:       0,
				PaginationLimit:    50,
				MaxPaginationLimit: 200,
			},
			wantErr: true,
			errMsg:  "max_key_length must be between 1 and 255",
		},
		{
			name: "max key length too large",
			config: Config{
				AllowedTypes:       []string{"post"},
				MaxKeyLength:       256,
				PaginationLimit:    50,
				MaxPaginationLimit: 200,
			},
			wantErr: true,
			errMsg:  "max_key_length must be between 1 and 255",
		},
		{
			name: "pagination limit too small",
			config: Config{
				AllowedTypes:       []string{"post"},
				MaxKeyLength:       255,
				PaginationLimit:    0,
				MaxPaginationLimit: 200,
			},
			wantErr: true,
			errMsg:  "pagination_limit must be between 1 and max_pagination_limit",
		},
		{
			name: "pagination limit too large",
			config: Config{
				AllowedTypes:       []string{"post"},
				MaxKeyLength:       255,
				PaginationLimit:    201,
				MaxPaginationLimit: 200,
			},
			wantErr: true,
			errMsg:  "pagination_limit must be between 1 and max_pagination_limit",
		},
		{
			name: "max pagination limit too large",
			config: Config{
				AllowedTypes:       []string{"post"},
				MaxKeyLength:       255,
				PaginationLimit:    50,
				MaxPaginationLimit: 1001,
			},
			wantErr: true,
			errMsg:  "max_pagination_limit must be between 1 and 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConfig_IsAllowedType(t *testing.T) {
	config := Config{
		AllowedTypes: []string{"post", "user", "product"},
	}

	tests := []struct {
		name         string
		resourceType string
		want         bool
	}{
		{
			name:         "allowed type - post",
			resourceType: "post",
			want:         true,
		},
		{
			name:         "allowed type - user",
			resourceType: "user",
			want:         true,
		},
		{
			name:         "not allowed type",
			resourceType: "comment",
			want:         false,
		},
		{
			name:         "empty string",
			resourceType: "",
			want:         false,
		},
		{
			name:         "case sensitive - Post",
			resourceType: "Post",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.IsAllowedType(tt.resourceType); got != tt.want {
				t.Errorf("IsAllowedType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if len(config.AllowedTypes) == 0 {
		t.Error("DefaultConfig() should have at least one allowed type")
	}

	if config.AllowedTypes[0] != "post" {
		t.Errorf("DefaultConfig() first allowed type = %v, want post", config.AllowedTypes[0])
	}

	if config.MaxKeyLength != 255 {
		t.Errorf("DefaultConfig() MaxKeyLength = %d, want 255", config.MaxKeyLength)
	}

	if config.OnlyPositiveValues != false {
		t.Errorf("DefaultConfig() OnlyPositiveValues = %v, want false", config.OnlyPositiveValues)
	}

	if config.PaginationLimit != 50 {
		t.Errorf("DefaultConfig() PaginationLimit = %d, want 50", config.PaginationLimit)
	}

	if config.MaxPaginationLimit != 200 {
		t.Errorf("DefaultConfig() MaxPaginationLimit = %d, want 200", config.MaxPaginationLimit)
	}

	if err := config.Validate(); err != nil {
		t.Errorf("DefaultConfig() should be valid, got error: %v", err)
	}
}
