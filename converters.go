package metrics

import (
	"github.com/google/uuid"
)

type MetricConverter struct{}

func (c *MetricConverter) CreateDTOToModel(dto MetricCreateDTO) Metric {
	return Metric{
		Id:         uuid.New().String(),
		Resource:   dto.Resource,
		ResourceId: dto.ResourceId,
		Key:        dto.Key,
		Value:      dto.Value,
	}
}

func (c *MetricConverter) UpdateDTOToModel(dto MetricUpdateDTO) Metric {
	return Metric{
		Value: dto.Value,
	}
}

func (c *MetricConverter) ModelToResponseDTO(model Metric) MetricResponseDTO {
	return MetricResponseDTO{
		ID:         model.Id,
		Resource:   model.Resource,
		ResourceID: model.ResourceId,
		Key:        model.Key,
		Value:      model.Value,
		CreatedAt:  model.CreatedAt,
	}
}

func (c *MetricConverter) ModelsToResponseDTOs(models []Metric) []MetricResponseDTO {
	dtos := make([]MetricResponseDTO, len(models))
	for i, model := range models {
		dtos[i] = c.ModelToResponseDTO(model)
	}
	return dtos
}
