package models

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/caraml-dev/merlin/config"
)

func TestParseGPUsConfig(t *testing.T) {
	type args struct {
		configGPUs []config.GPUConfig
	}
	tests := []struct {
		name string
		args args
		want GPUs
	}{
		{
			name: "successful parsing",
			args: args{
				configGPUs: []config.GPUConfig{
					{
						Values:       []string{"None", "1", "2"},
						DisplayName:  "NVIDIA P4",
						ResourceType: "nvidia.com/gpu",
						NodeSelector: map[string]string{
							"cloud.google.com/gke-accelerator": "nvidia-tesla-p4",
						},
						Tolerations: []corev1.Toleration{
							{
								Key:      "caraml/nvidia-tesla-p4",
								Operator: corev1.TolerationOpEqual,
								Value:    "enabled",
								Effect:   "NoSchedule",
							},
							{
								Key:      "nvidia.com/gpu",
								Operator: corev1.TolerationOpEqual,
								Value:    "present",
								Effect:   "NoSchedule",
							},
						},
						MinMonthlyCostPerGPU: 332.15,
						MaxMonthlyCostPerGPU: 332.15,
					},
				},
			},
			want: GPUs{
				{
					Values:       []string{"None", "1", "2"},
					DisplayName:  "NVIDIA P4",
					ResourceType: "nvidia.com/gpu",
					NodeSelector: map[string]string{
						"cloud.google.com/gke-accelerator": "nvidia-tesla-p4",
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "caraml/nvidia-tesla-p4",
							Operator: corev1.TolerationOpEqual,
							Value:    "enabled",
							Effect:   "NoSchedule",
						},
						{
							Key:      "nvidia.com/gpu",
							Operator: corev1.TolerationOpEqual,
							Value:    "present",
							Effect:   "NoSchedule",
						},
					},
					MinMonthlyCostPerGPU: 332.15,
					MaxMonthlyCostPerGPU: 332.15,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseGPUsConfig(tt.args.configGPUs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseGPUsConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
