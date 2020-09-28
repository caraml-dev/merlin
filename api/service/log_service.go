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
	"fmt"
	"io"
	"time"

	"github.com/gojek/merlin/models"
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type LogService interface {
	ReadLog(options *LogQuery) (io.ReadCloser, error)
}

type logService struct {
	// map of cluster name to cluster client
	logClients map[string]corev1.CoreV1Interface
}

// NewLogService create a log service
// logClients is a map of cluster name to its cluster client (CoreV1Interface)
func NewLogService(logClients map[string]corev1.CoreV1Interface) LogService {
	return &logService{logClients: logClients}
}

// ReadLog read log from a container with the given log options
// Return ReadCloser object which can be used to stream the log when the LogOptions.Follow is set to true
func (l logService) ReadLog(options *LogQuery) (io.ReadCloser, error) {
	logClient, ok := l.logClients[options.Cluster]
	if !ok {
		return nil, fmt.Errorf("unable to find cluster %s", options.Cluster)
	}

	return logClient.Pods(options.Namespace).GetLogs(options.PodName, options.ToKubernetesLogOption()).Stream()
}

type LogQuery struct {
	Name              string    `schema:"name,required"`
	PodName           string    `schema:"pod_name,required"`
	Namespace         string    `schema:"namespace,required"`
	Cluster           string    `schema:"cluster,required"`
	GcpProject        string    `schema:"gcp_project"`
	VersionEndpointId uuid.UUID `schema:"version_endpoint_id"`
	ModelId           models.Id `schema:"model_id"`

	Pretty       bool       `schema:"pretty"`
	Follow       bool       `schema:"follow"`
	Previous     bool       `schema:"previous"`
	SinceSeconds *int64     `schema:"since_seconds"`
	SinceTime    *time.Time `schema:"since_time"`
	Timestamps   bool       `schema:"timestamps"`
	TailLines    *int64     `schema:"tail_lines"`
	LimitBytes   *int64     `schema:"limit_bytes"`
}

func (opt *LogQuery) ToKubernetesLogOption() *v1.PodLogOptions {
	var sinceTime metav1.Time
	if opt.SinceTime != nil {
		sinceTime = metav1.NewTime(*opt.SinceTime)

	}
	return &v1.PodLogOptions{
		Container:    opt.Name,
		Follow:       opt.Follow,
		Previous:     opt.Previous,
		SinceSeconds: opt.SinceSeconds,
		SinceTime:    &sinceTime,
		Timestamps:   opt.Timestamps,
		TailLines:    opt.TailLines,
		LimitBytes:   opt.LimitBytes,
	}
}
