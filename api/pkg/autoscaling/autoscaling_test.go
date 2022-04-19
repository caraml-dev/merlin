package autoscaling

import (
	"testing"

	"github.com/gojek/merlin/pkg/deployment"
)

func TestValidateAutoscalingPolicy(t *testing.T) {
	type args struct {
		policy *AutoscalingPolicy
		mode   deployment.Mode
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "raw_deployment using cpu",
			args: args{
				policy: &AutoscalingPolicy{
					MetricsType: CPUUtilization,
					TargetValue: 10,
				},
				mode: deployment.RawDeploymentMode,
			},
			wantErr: false,
		},
		{
			name: "raw_deployment using cpu invalid value",
			args: args{
				policy: &AutoscalingPolicy{
					MetricsType: CPUUtilization,
					TargetValue: 110,
				},
				mode: deployment.RawDeploymentMode,
			},
			wantErr: true,
		},
		{
			name: "raw_deployment using memory",
			args: args{
				policy: &AutoscalingPolicy{
					MetricsType: MemoryUtilization,
					TargetValue: 10,
				},
				mode: deployment.RawDeploymentMode,
			},
			wantErr: true,
		},
		{
			name: "serverless using cpu",
			args: args{
				policy: &AutoscalingPolicy{
					MetricsType: CPUUtilization,
					TargetValue: 10,
				},
				mode: deployment.ServerlessDeploymentMode,
			},
			wantErr: false,
		},
		{
			name: "serverless using memory",
			args: args{
				policy: &AutoscalingPolicy{
					MetricsType: MemoryUtilization,
					TargetValue: 10,
				},
				mode: deployment.ServerlessDeploymentMode,
			},
			wantErr: false,
		},
		{
			name: "serverless using rps",
			args: args{
				policy: &AutoscalingPolicy{
					MetricsType: RPS,
					TargetValue: 10,
				},
				mode: deployment.ServerlessDeploymentMode,
			},
			wantErr: false,
		},
		{
			name: "serverless using concurrency",
			args: args{
				policy: &AutoscalingPolicy{
					MetricsType: Concurrency,
					TargetValue: 10,
				},
				mode: deployment.ServerlessDeploymentMode,
			},
			wantErr: false,
		},
		{
			name: "serverless using cpu invalid policy",
			args: args{
				policy: &AutoscalingPolicy{
					MetricsType: CPUUtilization,
					TargetValue: -1,
				},
				mode: deployment.ServerlessDeploymentMode,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateAutoscalingPolicy(tt.args.policy, tt.args.mode); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAutoscalingPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
