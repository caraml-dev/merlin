package utils

import "testing"

func TestLogTable(t *testing.T) {
	type args struct {
		headers []string
		rows    [][]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{},
			want: "",
		},
		{
			name: "empty isvc condition",
			args: args{
				headers: []string{"TYPE", "STATUS", "REASON", "MESSAGE"},
			},
			want: `┌──────┬────────┬────────┬─────────┐
│ TYPE │ STATUS │ REASON │ MESSAGE │
├──────┼────────┼────────┼─────────┤
└──────┴────────┴────────┴─────────┘`,
		},
		{
			name: "empty isvc predictor route not ready",
			args: args{
				headers: []string{"TYPE", "STATUS", "REASON", "MESSAGE"},
				rows: [][]string{
					{"LatestDeploymentReady", "Unknown", "PredictorConfigurationReady not ready", ""},
					{"PredictorConfigurationReady", "Unknown", "", ""},
					{"PredictorReady", "Unknown", "RevisionMissing", "Configuration \"custom-sleep-5-r1-predictor\" is waiting for a Revision to become ready."},
					{"PredictorRouteReady", "Unknown", "RevisionMissing", "Configuration \"custom-sleep-5-r1-predictor\" is waiting for a Revision to become ready."},
					{"Ready", "Unknown", "RevisionMissing", "Configuration \"custom-sleep-5-r1-predictor\" is waiting for a Revision to become ready."},
					{"RoutesReady", "Unknown", "PredictorRouteReady not ready", ""},
				},
			},
			want: `┌─────────────────────────────┬─────────┬───────────────────────────────────────┬────────────────────────────────────────────────────────────────────────────────────────┐
│ TYPE                        │ STATUS  │ REASON                                │ MESSAGE                                                                                │
├─────────────────────────────┼─────────┼───────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────┤
│ LatestDeploymentReady       │ Unknown │ PredictorConfigurationReady not ready │                                                                                        │
│ PredictorConfigurationReady │ Unknown │                                       │                                                                                        │
│ PredictorReady              │ Unknown │ RevisionMissing                       │ Configuration "custom-sleep-5-r1-predictor" is waiting for a Revision to become ready. │
│ PredictorRouteReady         │ Unknown │ RevisionMissing                       │ Configuration "custom-sleep-5-r1-predictor" is waiting for a Revision to become ready. │
│ Ready                       │ Unknown │ RevisionMissing                       │ Configuration "custom-sleep-5-r1-predictor" is waiting for a Revision to become ready. │
│ RoutesReady                 │ Unknown │ PredictorRouteReady not ready         │                                                                                        │
└─────────────────────────────┴─────────┴───────────────────────────────────────┴────────────────────────────────────────────────────────────────────────────────────────┘`,
		},
		{
			name: "isvc - image not found",
			args: args{
				headers: []string{"TYPE", "STATUS", "REASON", "MESSAGE"},
				rows: [][]string{
					{"LatestDeploymentReady", "False", "PredictorConfigurationReady not ready", ""},
					{"PredictorConfigurationReady", "False", "RevisionFailed", "Revision \"custom-sleep-6-r1-predictor-00001\" failed with message: Unable to fetch image \"asia.gcr.io/gcp-project/arief-test-pyfunc-sleepasdasdasd:1\": failed to resolve image to digest: HEAD https://asia.gcr.io/v2/gcp-project/arief-test-pyfunc-sleepasdasdasd/manifests/1: unexpected status code 404 Not Found (HEAD responses have no body, use GET for details)."},
					{"PredictorReady", "False", "RevisionMissing", "Configuration \"custom-sleep-6-r1-predictor\" does not have any ready Revision."},
					{"PredictorRouteReady", "False", "RevisionMissing", "Configuration \"custom-sleep-6-r1-predictor\" does not have any ready Revision."},
					{"Ready", "False", "RevisionMissing", "Configuration \"custom-sleep-6-r1-predictor\" does not have any ready Revision."},
					{"RoutesReady", "False", "PredictorRouteReady not ready", ""},
				},
			},
			want: `┌─────────────────────────────┬────────┬───────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ TYPE                        │ STATUS │ REASON                                │ MESSAGE                                                                                                                                                                                                                                                                                                                                                                 │
├─────────────────────────────┼────────┼───────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ LatestDeploymentReady       │ False  │ PredictorConfigurationReady not ready │                                                                                                                                                                                                                                                                                                                                                                         │
│ PredictorConfigurationReady │ False  │ RevisionFailed                        │ Revision "custom-sleep-6-r1-predictor-00001" failed with message: Unable to fetch image "asia.gcr.io/gcp-project/arief-test-pyfunc-sleepasdasdasd:1": failed to resolve image to digest: HEAD https://asia.gcr.io/v2/gcp-project/arief-test-pyfunc-sleepasdasdasd/manifests/1: unexpected status code 404 Not Found (HEAD responses have no body, use GET for details). │
│ PredictorReady              │ False  │ RevisionMissing                       │ Configuration "custom-sleep-6-r1-predictor" does not have any ready Revision.                                                                                                                                                                                                                                                                                           │
│ PredictorRouteReady         │ False  │ RevisionMissing                       │ Configuration "custom-sleep-6-r1-predictor" does not have any ready Revision.                                                                                                                                                                                                                                                                                           │
│ Ready                       │ False  │ RevisionMissing                       │ Configuration "custom-sleep-6-r1-predictor" does not have any ready Revision.                                                                                                                                                                                                                                                                                           │
│ RoutesReady                 │ False  │ PredictorRouteReady not ready         │                                                                                                                                                                                                                                                                                                                                                                         │
└─────────────────────────────┴────────┴───────────────────────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LogTable(tt.args.headers, tt.args.rows); got != tt.want {
				t.Errorf("LogTable() = %v, want %v", got, tt.want)
			}
		})
	}
}
