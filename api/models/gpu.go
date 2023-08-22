package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	corev1 "k8s.io/api/core/v1"

	"github.com/caraml-dev/merlin/config"
)

type GPU struct {
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
	// To deploy the models on a specific GPU node via taints and tolerations.
	Tolerations []corev1.Toleration `json:"tolerations"`
	// MinMonthlyCostPerGPU is the minimum monthly cost per GPU, for example, if you enable time-sharing GPUs with 8 max shared clients,
	// the minimum monthly cost per GPU is max_monthly_cost_per_gpu divided by 8.
	// MaxMonthlyCostPerGPU is the maximum monthly cost if you use the whole GPU.
	// https://cloud.google.com/compute/gpus-pricing#other-gpu-models
	MinMonthlyCostPerGPU float64 `json:"min_monthly_cost_per_gpu"`
	MaxMonthlyCostPerGPU float64 `json:"max_monthly_cost_per_gpu"`
}

type GPUs []GPU

func (gpus GPUs) Value() (driver.Value, error) {
	return json.Marshal(gpus)
}

func (gpus *GPUs) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &gpus)
}

// Function to parse []config.GPUConfig into models.GPU
func ParseGPUsConfig(configGPUs []config.GPUConfig) GPUs {
	gpus := []GPU{}

	for _, configGPU := range configGPUs {
		gpu := GPU{
			Values:               configGPU.Values,
			DisplayName:          configGPU.DisplayName,
			ResourceType:         configGPU.ResourceType,
			NodeSelector:         configGPU.NodeSelector,
			Tolerations:          configGPU.Tolerations,
			MinMonthlyCostPerGPU: configGPU.MinMonthlyCostPerGPU,
			MaxMonthlyCostPerGPU: configGPU.MaxMonthlyCostPerGPU,
		}
		gpus = append(gpus, gpu)
	}

	return gpus
}
