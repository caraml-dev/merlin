package cluster

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	metav1cfg "k8s.io/client-go/applyconfigurations/meta/v1"
	policyv1cfg "k8s.io/client-go/applyconfigurations/policy/v1"
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
