package cluster

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	metav1cfg "k8s.io/client-go/applyconfigurations/meta/v1"
	policyv1cfg "k8s.io/client-go/applyconfigurations/policy/v1"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/models"
)

func TestPodDisruptionBudget_BuildPDBSpec(t *testing.T) {
	defaultInt := 20
	defaultIntOrString := intstr.FromString("20%")
	defaultLabels := map[string]string{
		"gojek.com/app": "sklearn-sample-s",
	}

	type fields struct {
		Name                     string
		Namespace                string
		Labels                   map[string]string
		MaxUnavailablePercentage *int
		MinAvailablePercentage   *int
		Selector                 *metav1.LabelSelector
	}
	tests := []struct {
		name    string
		fields  fields
		want    *policyv1cfg.PodDisruptionBudgetSpecApplyConfiguration
		wantErr bool
	}{
		{
			name: "valid: enabled with min_available",
			fields: fields{
				Name:                     "sklearn-sample-s-1-model-pdb",
				Namespace:                "pdb-test",
				Labels:                   defaultLabels,
				MaxUnavailablePercentage: nil,
				MinAvailablePercentage:   &defaultInt,
			},
			want: &policyv1cfg.PodDisruptionBudgetSpecApplyConfiguration{
				MinAvailable: &defaultIntOrString,
				Selector: &metav1cfg.LabelSelectorApplyConfiguration{
					MatchLabels: defaultLabels,
				},
			},
			wantErr: false,
		},
		{
			name: "valid: enabled but both max_unavailable and min_available specified. will use min available",
			fields: fields{
				Name:                     "sklearn-sample-s-1-model-pdb",
				Namespace:                "pdb-test",
				Labels:                   defaultLabels,
				MaxUnavailablePercentage: &defaultInt,
				MinAvailablePercentage:   &defaultInt,
			},
			want: &policyv1cfg.PodDisruptionBudgetSpecApplyConfiguration{
				MinAvailable: &defaultIntOrString,
				Selector: &metav1cfg.LabelSelectorApplyConfiguration{
					MatchLabels: defaultLabels,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid: enabled but no max_unavailable and min_available",
			fields: fields{
				Name:                     "sklearn-sample-s-1-model-pdb",
				Namespace:                "pdb-test",
				Labels:                   map[string]string{},
				MaxUnavailablePercentage: nil,
				MinAvailablePercentage:   nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := PodDisruptionBudget{
				Name:                     tt.fields.Name,
				Namespace:                tt.fields.Namespace,
				Labels:                   tt.fields.Labels,
				MaxUnavailablePercentage: tt.fields.MaxUnavailablePercentage,
				MinAvailablePercentage:   tt.fields.MinAvailablePercentage,
			}
			got, err := cfg.BuildPDBSpec()
			if (err != nil) != tt.wantErr {
				t.Errorf("PodDisruptionBudget.BuildPDBSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodDisruptionBudget.BuildPDBSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreatePodDisruptionBudgets(t *testing.T) {
	twenty, eighty := 20, 80
	defaultLabels := map[string]string{
		"app":          "",
		"component":    "",
		"environment":  "",
		"orchestrator": "merlin",
		"stream":       "",
		"team":         "",
	}

	tests := map[string]struct {
		modelService *models.Service
		pdbConfig    config.PodDisruptionBudgetConfig
		expected     []*PodDisruptionBudget
	}{
		"bad pdb config": {
			modelService: &models.Service{
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 5,
				},
			},
			pdbConfig: config.PodDisruptionBudgetConfig{
				Enabled: true,
			},
			expected: []*PodDisruptionBudget{},
		},
		"model min replica low | minAvailablePercentage": {
			modelService: &models.Service{
				Name:      "test",
				Namespace: "test-ns",
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 1,
				},
				Transformer: &models.Transformer{
					Enabled: true,
					ResourceRequest: &models.ResourceRequest{
						MinReplica: 3,
					},
				},
			},
			pdbConfig: config.PodDisruptionBudgetConfig{
				Enabled:                true,
				MinAvailablePercentage: &twenty,
			},
			expected: []*PodDisruptionBudget{
				{
					Name:                   "test-transformer-pdb",
					Namespace:              "test-ns",
					Labels:                 defaultLabels,
					MinAvailablePercentage: &twenty,
				},
			},
		},
		"transformer min replica low | minAvailablePercentage": {
			modelService: &models.Service{
				Name:      "test",
				Namespace: "test-ns",
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 3,
				},
				Transformer: &models.Transformer{
					Enabled: true,
					ResourceRequest: &models.ResourceRequest{
						MinReplica: 1,
					},
				},
			},
			pdbConfig: config.PodDisruptionBudgetConfig{
				Enabled:                true,
				MinAvailablePercentage: &twenty,
			},
			expected: []*PodDisruptionBudget{
				{
					Name:                   "test-model-pdb",
					Namespace:              "test-ns",
					Labels:                 defaultLabels,
					MinAvailablePercentage: &twenty,
				},
			},
		},
		"all pdbs | minAvailablePercentage": {
			modelService: &models.Service{
				Name:      "test",
				Namespace: "test-ns",
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 5,
				},
				Transformer: &models.Transformer{
					Enabled: true,
					ResourceRequest: &models.ResourceRequest{
						MinReplica: 3,
					},
				},
			},
			pdbConfig: config.PodDisruptionBudgetConfig{
				Enabled:                true,
				MinAvailablePercentage: &twenty,
			},
			expected: []*PodDisruptionBudget{
				{
					Name:                   "test-model-pdb",
					Namespace:              "test-ns",
					Labels:                 defaultLabels,
					MinAvailablePercentage: &twenty,
				},
				{
					Name:                   "test-transformer-pdb",
					Namespace:              "test-ns",
					Labels:                 defaultLabels,
					MinAvailablePercentage: &twenty,
				},
			},
		},
		"all pdbs | maxUnavailablePercentage": {
			modelService: &models.Service{
				Name:      "test",
				Namespace: "test-ns",
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 5,
				},
				Transformer: &models.Transformer{
					Enabled: true,
					ResourceRequest: &models.ResourceRequest{
						MinReplica: 3,
					},
				},
			},
			pdbConfig: config.PodDisruptionBudgetConfig{
				Enabled:                  true,
				MaxUnavailablePercentage: &eighty,
			},
			expected: []*PodDisruptionBudget{
				{
					Name:                     "test-model-pdb",
					Namespace:                "test-ns",
					Labels:                   defaultLabels,
					MaxUnavailablePercentage: &eighty,
				},
				{
					Name:                     "test-transformer-pdb",
					Namespace:                "test-ns",
					Labels:                   defaultLabels,
					MaxUnavailablePercentage: &eighty,
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pdbs := createPodDisruptionBudgets(tt.modelService, tt.pdbConfig)
			assert.Equal(t, tt.expected, pdbs)
		})
	}
}
