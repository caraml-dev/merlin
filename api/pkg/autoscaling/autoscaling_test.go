package autoscaling

import (
	"testing"

	"github.com/gojek/merlin/pkg/deployment"
)

func TestValidateAutoscalingTarget(t *testing.T) {
	type args struct {
		target *AutoscalingTarget
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
				target: &AutoscalingTarget{
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
				target: &AutoscalingTarget{
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
				target: &AutoscalingTarget{
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
				target: &AutoscalingTarget{
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
				target: &AutoscalingTarget{
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
				target: &AutoscalingTarget{
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
				target: &AutoscalingTarget{
					MetricsType: Concurrency,
					TargetValue: 10,
				},
				mode: deployment.ServerlessDeploymentMode,
			},
			wantErr: false,
		},
		{
			name: "serverless using cpu invalid target",
			args: args{
				target: &AutoscalingTarget{
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
			if err := ValidateAutoscalingTarget(tt.args.target, tt.args.mode); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAutoscalingTarget() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
