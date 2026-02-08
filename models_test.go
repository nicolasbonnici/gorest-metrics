package metrics

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestMetric_TableName(t *testing.T) {
	metric := Metric{}
	if got := metric.TableName(); got != "metrics" {
		t.Errorf("TableName() = %v, want metrics", got)
	}
}

func TestCreateMetricRequest_Validate(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"post", "user"},
		MaxKeyLength:       255,
		OnlyPositiveValues: false,
	}

	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		req     *CreateMetricRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &CreateMetricRequest{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "view_count",
				Value:      100,
			},
			wantErr: false,
		},
		{
			name: "valid request with negative value",
			req: &CreateMetricRequest{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "delta",
				Value:      -50,
			},
			wantErr: false,
		},
		{
			name: "invalid UUID",
			req: &CreateMetricRequest{
				Resource:   "post",
				ResourceId: "not-a-uuid",
				Key:        "view_count",
				Value:      100,
			},
			wantErr: true,
			errMsg:  "resourceId must be a valid UUID",
		},
		{
			name: "resource not in allowed list",
			req: &CreateMetricRequest{
				Resource:   "comment",
				ResourceId: validUUID,
				Key:        "view_count",
				Value:      100,
			},
			wantErr: true,
			errMsg:  "resource type is not allowed",
		},
		{
			name: "empty key",
			req: &CreateMetricRequest{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "   ",
				Value:      100,
			},
			wantErr: true,
			errMsg:  "key cannot be empty",
		},
		{
			name: "key exceeds max length",
			req: &CreateMetricRequest{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        strings.Repeat("a", 256),
				Value:      100,
			},
			wantErr: true,
			errMsg:  "key exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(config)
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

func TestCreateMetricRequest_Validate_OnlyPositiveValues(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"post"},
		MaxKeyLength:       255,
		OnlyPositiveValues: true,
	}

	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		req     *CreateMetricRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "positive value allowed",
			req: &CreateMetricRequest{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "view_count",
				Value:      100,
			},
			wantErr: false,
		},
		{
			name: "zero value allowed",
			req: &CreateMetricRequest{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "view_count",
				Value:      0,
			},
			wantErr: false,
		},
		{
			name: "negative value rejected",
			req: &CreateMetricRequest{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "view_count",
				Value:      -1,
			},
			wantErr: true,
			errMsg:  "value must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(config)
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

func TestUpdateMetricRequest_Validate(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"post"},
		MaxKeyLength:       255,
		OnlyPositiveValues: false,
	}

	tests := []struct {
		name    string
		req     *UpdateMetricRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid positive value",
			req: &UpdateMetricRequest{
				Value: 200,
			},
			wantErr: false,
		},
		{
			name: "valid negative value",
			req: &UpdateMetricRequest{
				Value: -100,
			},
			wantErr: false,
		},
		{
			name: "valid zero value",
			req: &UpdateMetricRequest{
				Value: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(config)
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

func TestUpdateMetricRequest_Validate_OnlyPositiveValues(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"post"},
		MaxKeyLength:       255,
		OnlyPositiveValues: true,
	}

	tests := []struct {
		name    string
		req     *UpdateMetricRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "positive value allowed",
			req: &UpdateMetricRequest{
				Value: 100,
			},
			wantErr: false,
		},
		{
			name: "negative value rejected",
			req: &UpdateMetricRequest{
				Value: -1,
			},
			wantErr: true,
			errMsg:  "value must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(config)
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
