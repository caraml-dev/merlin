package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/caraml-dev/merlin/config"
)

type Gpu struct {
	// Values limits how many GPUs can be requested by users.
	// Example: "none", "1", "2", "4"
	Values []string `json:"values"`
	// Specifies how the accelerator type will be written in the UI.
	// Example: "NVIDIA T4"
	DisplayName string `json:"display_name"`
	// Specifies how the accelerator type will be translated to
	// K8s resource type. Example: nvidia.com/gpu
	ResourceType string `json:"resource_type"`
	// To deploy the models on a specific GPU node.
	NodeSelector map[string]string `json:"node_selector"`
	// https://cloud.google.com/compute/gpus-pricing#other-gpu-models
	MonthlyCostPerGpu float64 `json:"monthly_cost_per_gpu"`
}

type Gpus []Gpu

func (gpus Gpus) Value() (driver.Value, error) {
	return json.Marshal(gpus)
}

func (gpus *Gpus) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &gpus)
}

// Function to parse []config.GpuConfig into models.Gpu
func ParseGpusConfig(configGpus []config.GpuConfig) Gpus {
	gpus := []Gpu{}

	for _, configGpu := range configGpus {
		gpu := Gpu{
			Values:            configGpu.Values,
			DisplayName:       configGpu.DisplayName,
			ResourceType:      configGpu.ResourceType,
			NodeSelector:      configGpu.NodeSelector,
			MonthlyCostPerGpu: configGpu.MonthlyCostPerGpu,
		}
		gpus = append(gpus, gpu)
	}

	return gpus
}
