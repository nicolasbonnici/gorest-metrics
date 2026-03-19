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

func TestCreateMetricDTO_Validate(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"post", "user"},
		MaxKeyLength:       255,
		OnlyPositiveValues: false,
	}

	hooks := NewMetricHooks(config)
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		dto     MetricCreateDTO
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			dto: MetricCreateDTO{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "view_count",
				Value:      100,
			},
			wantErr: false,
		},
		{
			name: "valid request with negative value",
			dto: MetricCreateDTO{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "delta",
				Value:      -50,
			},
			wantErr: false,
		},
		{
			name: "invalid UUID",
			dto: MetricCreateDTO{
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
			dto: MetricCreateDTO{
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
			dto: MetricCreateDTO{
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
			dto: MetricCreateDTO{
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
			model := &Metric{}
			err := hooks.CreateHook(nil, tt.dto, model)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateHook() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CreateHook() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("CreateHook() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestCreateMetricDTO_Validate_OnlyPositiveValues(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"post"},
		MaxKeyLength:       255,
		OnlyPositiveValues: true,
	}

	hooks := NewMetricHooks(config)
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		dto     MetricCreateDTO
		wantErr bool
		errMsg  string
	}{
		{
			name: "positive value allowed",
			dto: MetricCreateDTO{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "view_count",
				Value:      100,
			},
			wantErr: false,
		},
		{
			name: "zero value allowed",
			dto: MetricCreateDTO{
				Resource:   "post",
				ResourceId: validUUID,
				Key:        "view_count",
				Value:      0,
			},
			wantErr: false,
		},
		{
			name: "negative value rejected",
			dto: MetricCreateDTO{
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
			model := &Metric{}
			err := hooks.CreateHook(nil, tt.dto, model)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateHook() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CreateHook() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("CreateHook() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestUpdateMetricDTO_Validate(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"post"},
		MaxKeyLength:       255,
		OnlyPositiveValues: false,
	}

	hooks := NewMetricHooks(config)

	tests := []struct {
		name    string
		dto     MetricUpdateDTO
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid positive value",
			dto: MetricUpdateDTO{
				Value: 200,
			},
			wantErr: false,
		},
		{
			name: "valid negative value",
			dto: MetricUpdateDTO{
				Value: -100,
			},
			wantErr: false,
		},
		{
			name: "valid zero value",
			dto: MetricUpdateDTO{
				Value: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &Metric{}
			err := hooks.UpdateHook(nil, tt.dto, model)
			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateHook() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("UpdateHook() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("UpdateHook() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestUpdateMetricDTO_Validate_OnlyPositiveValues(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"post"},
		MaxKeyLength:       255,
		OnlyPositiveValues: true,
	}

	hooks := NewMetricHooks(config)

	tests := []struct {
		name    string
		dto     MetricUpdateDTO
		wantErr bool
		errMsg  string
	}{
		{
			name: "positive value allowed",
			dto: MetricUpdateDTO{
				Value: 100,
			},
			wantErr: false,
		},
		{
			name: "negative value rejected",
			dto: MetricUpdateDTO{
				Value: -1,
			},
			wantErr: true,
			errMsg:  "value must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &Metric{}
			err := hooks.UpdateHook(nil, tt.dto, model)
			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateHook() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("UpdateHook() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("UpdateHook() unexpected error = %v", err)
				}
			}
		})
	}
}
