package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/caraml-dev/merlin/models"
)

func int64Ptr(v int64) *int64 { return &v }

func TestValidateTolerations(t *testing.T) {
	tests := []struct {
		name        string
		tolerations []corev1.Toleration
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil tolerations — always valid",
			tolerations: nil,
			wantErr:     false,
		},
		{
			name:        "empty tolerations — always valid",
			tolerations: []corev1.Toleration{},
			wantErr:     false,
		},
		{
			name: "valid Equal toleration with NoSchedule",
			tolerations: []corev1.Toleration{
				{Key: "dedicated", Operator: corev1.TolerationOpEqual, Value: "ml-team", Effect: corev1.TaintEffectNoSchedule},
			},
			wantErr: false,
		},
		{
			name: "valid Exists toleration with empty value and NoExecute",
			tolerations: []corev1.Toleration{
				{Key: "spot", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoExecute},
			},
			wantErr: false,
		},
		{
			name: "valid empty operator defaults to Equal",
			tolerations: []corev1.Toleration{
				{Key: "key", Operator: "", Value: "val", Effect: corev1.TaintEffectPreferNoSchedule},
			},
			wantErr: false,
		},
		{
			name: "valid empty effect matches all effects",
			tolerations: []corev1.Toleration{
				{Key: "key", Operator: corev1.TolerationOpEqual, Value: "val", Effect: ""},
			},
			wantErr: false,
		},
		{
			name: "valid TolerationSeconds on NoExecute",
			tolerations: []corev1.Toleration{
				{Key: "key", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoExecute, TolerationSeconds: int64Ptr(300)},
			},
			wantErr: false,
		},
		{
			name: "multiple valid tolerations",
			tolerations: []corev1.Toleration{
				{Key: "key1", Operator: corev1.TolerationOpEqual, Value: "v1", Effect: corev1.TaintEffectNoSchedule},
				{Key: "key2", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoExecute, TolerationSeconds: int64Ptr(60)},
			},
			wantErr: false,
		},
		// --- invalid cases ---
		{
			name: "invalid effect",
			tolerations: []corev1.Toleration{
				{Key: "key", Operator: corev1.TolerationOpEqual, Value: "val", Effect: "InvalidEffect"},
			},
			wantErr:     true,
			errContains: "invalid effect",
		},
		{
			name: "invalid operator",
			tolerations: []corev1.Toleration{
				{Key: "key", Operator: "NotAValidOperator", Value: "val", Effect: corev1.TaintEffectNoSchedule},
			},
			wantErr:     true,
			errContains: "invalid operator",
		},
		{
			name: "Exists operator with a non-empty value",
			tolerations: []corev1.Toleration{
				{Key: "key", Operator: corev1.TolerationOpExists, Value: "should-be-empty", Effect: corev1.TaintEffectNoSchedule},
			},
			wantErr:     true,
			errContains: "must not specify a value",
		},
		{
			name: "TolerationSeconds on NoSchedule (only valid for NoExecute)",
			tolerations: []corev1.Toleration{
				{Key: "key", Operator: corev1.TolerationOpEqual, Value: "v", Effect: corev1.TaintEffectNoSchedule, TolerationSeconds: int64Ptr(120)},
			},
			wantErr:     true,
			errContains: "tolerationSeconds is only valid for effect 'NoExecute'",
		},
		{
			name: "TolerationSeconds on PreferNoSchedule",
			tolerations: []corev1.Toleration{
				{Key: "key", Operator: corev1.TolerationOpEqual, Value: "v", Effect: corev1.TaintEffectPreferNoSchedule, TolerationSeconds: int64Ptr(60)},
			},
			wantErr:     true,
			errContains: "tolerationSeconds is only valid for effect 'NoExecute'",
		},
		{
			name: "second toleration in list is invalid",
			tolerations: []corev1.Toleration{
				{Key: "ok", Operator: corev1.TolerationOpEqual, Value: "v", Effect: corev1.TaintEffectNoSchedule},
				{Key: "bad", Operator: corev1.TolerationOpExists, Value: "oops", Effect: corev1.TaintEffectNoSchedule},
			},
			wantErr:     true,
			errContains: "must not specify a value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTolerations(tt.tolerations)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceRequestValidation_Tolerations(t *testing.T) {
	tests := []struct {
		name    string
		req     *models.ResourceRequest
		trans   *models.Transformer
		wantErr bool
	}{
		{
			name: "nil resource request passes",
			req:  nil,
		},
		{
			name: "valid resource request with tolerations passes",
			req: &models.ResourceRequest{
				MinReplica: 1, MaxReplica: 2,
				Tolerations: []corev1.Toleration{
					{Key: "dedicated", Operator: corev1.TolerationOpEqual, Value: "ml", Effect: corev1.TaintEffectNoSchedule},
				},
			},
		},
		{
			name: "valid transformer resource request with tolerations passes",
			trans: &models.Transformer{
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 1, MaxReplica: 2,
					Tolerations: []corev1.Toleration{
						{Key: "dedicated", Operator: corev1.TolerationOpEqual, Value: "transformer", Effect: corev1.TaintEffectNoSchedule},
					},
				},
			},
		},
		{
			name: "resource request with invalid toleration operator fails",
			req: &models.ResourceRequest{
				MinReplica: 1, MaxReplica: 2,
				Tolerations: []corev1.Toleration{
					{Key: "k", Operator: "BadOp", Value: "v"},
				},
			},
			wantErr: true,
		},
		{
			name: "transformer resource request with invalid toleration operator fails",
			trans: &models.Transformer{
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 1, MaxReplica: 2,
					Tolerations: []corev1.Toleration{
						{Key: "k", Operator: "BadOp", Value: "v"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := &models.VersionEndpoint{
				ResourceRequest: tt.req,
				Transformer:     tt.trans,
			}
			validator := resourceRequestValidation(endpoint)
			err := validator.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceRequestValidation_TransformerTolerations(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    *models.VersionEndpoint
		wantErr     bool
		errContains string
	}{
		{
			name:     "nil transformer passes",
			endpoint: &models.VersionEndpoint{},
		},
		{
			name: "transformer with nil resource request passes",
			endpoint: &models.VersionEndpoint{
				Transformer: &models.Transformer{Enabled: true},
			},
		},
		{
			name: "transformer with valid tolerations passes",
			endpoint: &models.VersionEndpoint{
				Transformer: &models.Transformer{
					Enabled: true,
					ResourceRequest: &models.ResourceRequest{
						MinReplica: 1, MaxReplica: 2,
						Tolerations: []corev1.Toleration{
							{Key: "dedicated", Operator: corev1.TolerationOpEqual, Value: "ml", Effect: corev1.TaintEffectNoSchedule},
						},
					},
				},
			},
		},
		{
			name: "transformer with invalid toleration fails even when predictor resource request is nil",
			endpoint: &models.VersionEndpoint{
				Transformer: &models.Transformer{
					Enabled: true,
					ResourceRequest: &models.ResourceRequest{
						MinReplica: 1, MaxReplica: 2,
						Tolerations: []corev1.Toleration{
							{Key: "k", Operator: corev1.TolerationOpExists, Value: "should-be-empty", Effect: corev1.TaintEffectNoSchedule},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "transformer resource request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := resourceRequestValidation(tt.endpoint)
			err := validator.validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
