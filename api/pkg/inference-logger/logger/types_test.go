package logger

import (
	"reflect"
	"testing"
)

func TestParseSinkKindAndUrl(t *testing.T) {
	type args struct {
		logUrl string
	}
	tests := []struct {
		name  string
		args  args
		want  LoggerSinkKind
		want1 string
	}{
		{
			"kafka",
			args{
				logUrl: "kafka:localhost:9092",
			},
			Kafka,
			"localhost:9092",
		},
		{
			"newrelic",
			args{
				logUrl: "newrelic:https://log-api.newrelic.com/log/v1?Api-Key=LICENSEKEY",
			},
			NewRelic,
			"https://log-api.newrelic.com/log/v1?Api-Key=LICENSEKEY",
		},
		{
			"console",
			args{
				logUrl: "console:localhost:8080",
			},
			Console,
			"localhost:8080",
		},
		{
			"invalid, fallback to console",
			args{
				logUrl: "invalid:localhost:8080",
			},
			Console,
			"invalid:localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ParseSinkKindAndUrl(tt.args.logUrl)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSinkKindAndUrl() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseSinkKindAndUrl() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
