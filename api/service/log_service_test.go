// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/caraml-dev/merlin/cluster"
	"github.com/caraml-dev/merlin/cluster/mocks"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLogLine_generateText(t *testing.T) {
	color.NoColor = false

	type fields struct {
		Timestamp     time.Time
		Namespace     string
		PodName       string
		ContainerName string
		TextPayload   string
		PrefixColor   *color.Color
	}
	type args struct {
		options LogQuery
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			"no prefix",
			fields{
				Timestamp:   now,
				TextPayload: "hello",
			},
			args{},
			nowRFC3339 + " hello\n",
		},
		{
			"pod and container prefix",
			fields{
				Timestamp:   now,
				TextPayload: "hello",

				PodName:       "pod",
				ContainerName: "container",
				PrefixColor:   color.New(color.FgRed),
			},
			args{
				LogQuery{
					Prefix: prefixAddPodContainer,
				},
			},
			"\x1b[31mpod\x1b[0m \x1b[31mcontainer\x1b[0m " + nowRFC3339 + " hello\n",
		},
		{
			"pod prefix",
			fields{
				Timestamp:   now,
				TextPayload: "hello",

				PodName:       "pod",
				ContainerName: "container",
				PrefixColor:   color.New(color.FgRed),
			},
			args{
				LogQuery{
					Prefix: prefixAddPod,
				},
			},
			"\x1b[31mpod\x1b[0m " + nowRFC3339 + " hello\n",
		},
		{
			"container prefix",
			fields{
				Timestamp:   now,
				TextPayload: "hello",

				PodName:       "pod",
				ContainerName: "container",
				PrefixColor:   color.New(color.FgRed),
			},
			args{
				LogQuery{
					Prefix: prefixAddContainer,
				},
			},
			"\x1b[31mcontainer\x1b[0m " + nowRFC3339 + " hello\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LogLine{
				Timestamp:     tt.fields.Timestamp,
				Namespace:     tt.fields.Namespace,
				PodName:       tt.fields.PodName,
				ContainerName: tt.fields.ContainerName,
				TextPayload:   tt.fields.TextPayload,
				PrefixColor:   tt.fields.PrefixColor,
			}
			if got := l.generateText(tt.args.options); got != tt.want {
				t.Errorf("LogLine.generateText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_logService_StreamLogs(t *testing.T) {
	logLineCh := make(chan string)
	stopCh := make(chan struct{})
	options := &LogQuery{
		ProjectName: "test-project",
		ModelID:     "1",
		ModelName:   "test-model",
		VersionID:   "1",
		RevisionID:  2,

		Cluster:       "test-cluster",
		Namespace:     "test-namespace",
		ComponentType: "model",

		Prefix: prefixAddPodContainer,
	}

	mockController := &mocks.Controller{}
	mockControllers := map[string]cluster.Controller{
		"test-cluster": mockController,
	}

	mockController.On("ListPods", context.Background(), "test-namespace", "component=predictor,serving.kserve.io/inferenceservice in (test-model-1,test-model-1-r1,test-model-1-r2)").
		Return(&v1.PodList{
			Items: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-model-1-r2-predictor-a",
						Labels: map[string]string{
							"component":                          "predictor",
							"serving.kserve.io/inferenceservice": "test-model-1-r2",
						},
					},
					Spec: v1.PodSpec{
						InitContainers: []v1.Container{
							{Name: "storage-initializer"},
						},
						Containers: []v1.Container{
							{Name: "kfserving-container"},
							{Name: "queue-proxy"},
							{Name: "inferenceservice-logger"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-model-1-r2-predictor-b",
						Labels: map[string]string{
							"component":                          "predictor",
							"serving.kserve.io/inferenceservice": "test-model-1-r2",
						},
					},
					Spec: v1.PodSpec{
						InitContainers: []v1.Container{
							{Name: "storage-initializer"},
						},
						Containers: []v1.Container{
							{Name: "kfserving-container"},
							{Name: "queue-proxy"},
							{Name: "inferenceservice-logger"},
						},
					},
				},
			},
		}, nil)

	pods := []string{"test-model-1-r2-predictor-a", "test-model-1-r2-predictor-b"}
	containers := []string{"storage-initializer", "kfserving-container", "inferenceservice-logger"}

	for _, pod := range pods {
		for _, container := range containers {
			r := io.NopCloser(strings.NewReader(nowRFC3339 + " log from " + pod + "/" + container))

			c := container
			mockController.On("StreamPodLogs", context.Background(), "test-namespace", pod, mock.MatchedBy(func(opts *v1.PodLogOptions) bool {
				return opts.Container == c
			})).
				Return(r, nil)
		}
	}

	got := []string{}

	go func() {
		for logLine := range logLineCh {
			got = append(got, logLine)
			if len(got) == len(pods)*len(containers) {
				close(stopCh)
				return
			}
		}
	}()

	l := NewLogService(mockControllers)
	err := l.StreamLogs(context.Background(), logLineCh, stopCh, options)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(got))
}

func Test_logService_getLabelSelector(t *testing.T) {
	type fields struct {
		clusterControllers map[string]cluster.Controller
	}
	type args struct {
		query LogQuery
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "model without revision",
			fields: fields{},
			args: args{query: LogQuery{
				ComponentType: "model",
				ProjectName:   "test-project",
				ModelName:     "test-model",
				VersionID:     "1",
				RevisionID:    0,
			}},
			want: "component=predictor,serving.kserve.io/inferenceservice in (test-model-1)",
		},
		{
			name:   "model name with number without revision",
			fields: fields{},
			args: args{query: LogQuery{
				ComponentType: "model",
				ProjectName:   "test-project",
				ModelName:     "test-model-0-1-2",
				VersionID:     "10",
				RevisionID:    0,
			}},
			want: "component=predictor,serving.kserve.io/inferenceservice in (test-model-0-1-2-10)",
		},
		{
			name:   "model with revision",
			fields: fields{},
			args: args{query: LogQuery{
				ComponentType: "model",
				ProjectName:   "test-project",
				ModelName:     "test-model",
				VersionID:     "1",
				RevisionID:    3,
			}},
			want: "component=predictor,serving.kserve.io/inferenceservice in (test-model-1,test-model-1-r1,test-model-1-r2,test-model-1-r3)",
		},
		{
			name:   "model name with number with revision",
			fields: fields{},
			args: args{query: LogQuery{
				ComponentType: "model",
				ProjectName:   "test-project",
				ModelName:     "test-model-0-1-2",
				VersionID:     "10",
				RevisionID:    2,
			}},
			want: "component=predictor,serving.kserve.io/inferenceservice in (test-model-0-1-2-10,test-model-0-1-2-10-r1,test-model-0-1-2-10-r2)",
		},
		{
			name:   "transformer without revision",
			fields: fields{},
			args: args{query: LogQuery{
				ComponentType: "transformer",
				ProjectName:   "test-project",
				ModelName:     "test-model",
				VersionID:     "1",
				RevisionID:    0,
			}},
			want: "component=transformer,serving.kserve.io/inferenceservice in (test-model-1)",
		},
		{
			name:   "transformer name with number without revision",
			fields: fields{},
			args: args{query: LogQuery{
				ComponentType: "transformer",
				ProjectName:   "test-project",
				ModelName:     "test-model-0-1-2",
				VersionID:     "10",
				RevisionID:    0,
			}},
			want: "component=transformer,serving.kserve.io/inferenceservice in (test-model-0-1-2-10)",
		},
		{
			name:   "transformer with revision",
			fields: fields{},
			args: args{query: LogQuery{
				ComponentType: "transformer",
				ProjectName:   "test-project",
				ModelName:     "test-model",
				VersionID:     "1",
				RevisionID:    3,
			}},
			want: "component=transformer,serving.kserve.io/inferenceservice in (test-model-1,test-model-1-r1,test-model-1-r2,test-model-1-r3)",
		},
		{
			name:   "transformer name with number with revision",
			fields: fields{},
			args: args{query: LogQuery{
				ComponentType: "transformer",
				ProjectName:   "test-project",
				ModelName:     "test-model-0-1-2",
				VersionID:     "10",
				RevisionID:    2,
			}},
			want: "component=transformer,serving.kserve.io/inferenceservice in (test-model-0-1-2-10,test-model-0-1-2-10-r1,test-model-0-1-2-10-r2)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logService{
				clusterControllers: tt.fields.clusterControllers,
			}
			if got := l.getLabelSelector(tt.args.query); got != tt.want {
				t.Errorf("logService.getLabelSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}
