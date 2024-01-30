package utils

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestParsePodContainerStatuses(t *testing.T) {
	type args struct {
		containerStatuses []v1.ContainerStatus
	}
	tests := []struct {
		name         string
		args         args
		wantPodTable string
		wantMessage  string
		wantReason   string
	}{
		{
			name: "empty",
			args: args{
				containerStatuses: []v1.ContainerStatus{},
			},
			wantPodTable: "",
			wantMessage:  "",
			wantReason:   "",
		},
		{
			name: "waiting",
			args: args{
				containerStatuses: []v1.ContainerStatus{
					{
						Name: "test",
						State: v1.ContainerState{
							Waiting: &v1.ContainerStateWaiting{
								Reason: "test",
							},
						},
					},
				},
			},
			wantPodTable: `┌────────────────┬─────────┬───────────┬────────┐
│ CONTAINER NAME │ STATUS  │ EXIT CODE │ REASON │
├────────────────┼─────────┼───────────┼────────┤
│ test           │ Waiting │ -         │ test   │
└────────────────┴─────────┴───────────┴────────┘`,
			wantMessage: "",
			wantReason:  "",
		},
		{
			name: "terminated",
			args: args{
				containerStatuses: []v1.ContainerStatus{
					{
						Name: "test",
						State: v1.ContainerState{
							Terminated: &v1.ContainerStateTerminated{
								ExitCode: 1,
								Reason:   "test",
								Message:  "test",
							},
						},
					},
				},
			},
			wantPodTable: `┌────────────────┬────────────┬───────────┬────────┐
│ CONTAINER NAME │ STATUS     │ EXIT CODE │ REASON │
├────────────────┼────────────┼───────────┼────────┤
│ test           │ Terminated │ 1         │ test   │
└────────────────┴────────────┴───────────┴────────┘`,
			wantMessage: "test",
			wantReason:  "test",
		},
		{
			name: "running",
			args: args{
				containerStatuses: []v1.ContainerStatus{
					{
						Name: "test",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
			wantPodTable: "",
			wantMessage:  "",
			wantReason:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPodTable, gotMessage, gotReason := ParsePodContainerStatuses(tt.args.containerStatuses)
			if gotPodTable != tt.wantPodTable {
				t.Errorf("ParsePodContainerStatuses() gotPodTable = %v, want %v", gotPodTable, tt.wantPodTable)
			}
			if gotMessage != tt.wantMessage {
				t.Errorf("ParsePodContainerStatuses() gotMessage = %v, want %v", gotMessage, tt.wantMessage)
			}
			if gotReason != tt.wantReason {
				t.Errorf("ParsePodContainerStatuses() gotReason = %v, want %v", gotReason, tt.wantReason)
			}
		})
	}
}
