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
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	KnativeServiceLabelKey = "serving.knative.dev/service"
)

var (
	skippedContainers = []string{"queue-proxy"}
)

// PodLog represents a single log line from a container running in Kubernetes.
type PodLog struct {
	// Log timestamp in RFC3339 format
	Timestamp time.Time `json:"timestamp"`
	// Kubernetes namespace where the pod running the container is created
	Namespace string `json:"namespace"`
	// Pod name running the container that produces this log
	PodName string `json:"pod_name"`
	// Container name that produces this log
	ContainerName string `json:"container_name"`
	// Log in text format
	TextPayload string `json:"text_payload,omitempty"`
}

type ReadLogStream struct {
	Stream        io.ReadCloser
	PodName       string
	ContainerName string
}

type LogService interface {
	StreamLogs(logs chan PodLog, stopCh chan struct{}, options *LogQuery) error
}

type logService struct {
	// map of cluster name to cluster client
	clusterClients map[string]corev1.CoreV1Interface
}

// NewLogService create a log service
// clusterClients is a map of cluster name to its cluster client (CoreV1Interface)
func NewLogService(clusterClients map[string]corev1.CoreV1Interface) LogService {
	return &logService{clusterClients: clusterClients}
}

func (l logService) StreamLogs(podLogs chan PodLog, stopCh chan struct{}, options *LogQuery) error {
	clusterClient, ok := l.clusterClients[options.Cluster]
	if !ok {
		return fmt.Errorf("unable to find cluster %s", options.Cluster)
	}

	namespace := options.Namespace

	labelSelector := l.getLabelSelector(*options)
	pods, err := clusterClient.Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}

	go func() {
		for {
			allPodLogs := make([]*PodLog, 0)

			for _, pod := range pods.Items {
				for _, initContainer := range pod.Spec.InitContainers {
					options.ContainerName = initContainer.Name
					stream, err := clusterClient.Pods(namespace).GetLogs(pod.Name, options.ToKubernetesLogOption()).Stream()
					if err != nil {
						// Error is handled here by logging it rather than returned because the caller usually does not know how to
						// handle it. Example of what can trigger ListPodLogs error: while the container is being created/terminated
						// Kubernetes API server will return error when logs are requested. In such case, it is better to return
						// empty logs and let the caller retry after the container becomes ready eventually.
						log.Warnf("Failed to ListPodLogs: %s", err.Error())
						return
					}

					scanner := bufio.NewScanner(stream)
					podLogs := make([]*PodLog, 0)
					for scanner.Scan() {
						logLine := scanner.Text()

						// A log line from Kubernetes API server will follow this format:
						// 2020-07-14T07:48:14.191189249Z {"msg":"log message"}
						timestampIndex := strings.Index(logLine, " ")
						if timestampIndex < 0 {
							// Missing expected RFC3339 timstamp in the log line, skip to next line
							continue
						}
						if (len(logLine) - 1) <= timestampIndex {
							// Empty log message, skip to next log line
							continue
						}

						timestamp, err := time.Parse(time.RFC3339, logLine[:timestampIndex])
						if err != nil {
							log.Warnf("log message timestamp is not in RFC3339 format: %s", logLine[:timestampIndex])
							// Log timestamp value from Kube API server has invalid format, skip to next line
							continue
						}

						// We require this check because we send (SinceTime - 1sec) to Kube API Server
						if options.SinceTime != nil && (timestamp == *options.SinceTime || timestamp.Before(*options.SinceTime)) {
							continue
						}

						log := &PodLog{
							Timestamp:     timestamp,
							Namespace:     namespace,
							PodName:       pod.Name,
							ContainerName: options.ContainerName,
							TextPayload:   logLine[timestampIndex+1:],
						}

						podLogs = append(podLogs, log)
					}

					allPodLogs = append(allPodLogs, podLogs...)
				}

				for _, container := range pod.Spec.Containers {
					for _, skippedContainer := range skippedContainers {
						if container.Name == skippedContainer {
							break
						}

						options.ContainerName = container.Name
						stream, err := clusterClient.Pods(namespace).GetLogs(pod.Name, options.ToKubernetesLogOption()).Stream()
						if err != nil {
							// Error is handled here by logging it rather than returned because the caller usually does not know how to
							// handle it. Example of what can trigger ListPodLogs error: while the container is being created/terminated
							// Kubernetes API server will return error when logs are requested. In such case, it is better to return
							// empty logs and let the caller retry after the container becomes ready eventually.
							log.Warnf("Failed to ListPodLogs: %s", err.Error())
							return
						}

						scanner := bufio.NewScanner(stream)
						podLogs := make([]*PodLog, 0)
						for scanner.Scan() {
							logLine := scanner.Text()

							// A log line from Kubernetes API server will follow this format:
							// 2020-07-14T07:48:14.191189249Z {"msg":"log message"}
							timestampIndex := strings.Index(logLine, " ")
							if timestampIndex < 0 {
								// Missing expected RFC3339 timstamp in the log line, skip to next line
								continue
							}
							if (len(logLine) - 1) <= timestampIndex {
								// Empty log message, skip to next log line
								continue
							}

							timestamp, err := time.Parse(time.RFC3339, logLine[:timestampIndex])
							if err != nil {
								log.Warnf("log message timestamp is not in RFC3339 format: %s", logLine[:timestampIndex])
								// Log timestamp value from Kube API server has invalid format, skip to next line
								continue
							}

							// We require this check because we send (SinceTime - 1sec) to Kube API Server
							if options.SinceTime != nil && (timestamp == *options.SinceTime || timestamp.Before(*options.SinceTime)) {
								continue
							}

							log := &PodLog{
								Timestamp:     timestamp,
								Namespace:     namespace,
								PodName:       pod.Name,
								ContainerName: options.ContainerName,
								TextPayload:   logLine[timestampIndex+1:],
							}

							podLogs = append(podLogs, log)
						}

						allPodLogs = append(allPodLogs, podLogs...)
					}
				}
			}

			// Sort all the logs by timestamp in ascending order
			sort.Slice(allPodLogs, func(i, j int) bool {
				return allPodLogs[i].Timestamp.Before(allPodLogs[j].Timestamp)
			})

			for _, podLog := range allPodLogs {
				podLogs <- *podLog
			}

			now := time.Now()
			options.SinceTime = &now

			time.Sleep(1 * time.Second)
		}
	}()

	<-stopCh
	return nil
}

func (l logService) getLabelSelector(query LogQuery) string {
	switch query.ComponentType {
	case models.ModelComponentType:
		return KnativeServiceLabelKey + "=" + query.ModelName + "-" + query.VersionID + "-predictor-default"
	case models.TransformerComponentType:
		return KnativeServiceLabelKey + "=" + query.ModelName + "-" + query.VersionID + "-transformer-default"
	}
	return ""
}

type LogQuery struct {
	ModelID         string `schema:"model_id"`
	ModelName       string `schema:"model_name"`
	VersionID       string `schema:"version_id"`
	PredictionJobID string `schema:"prediction_job_id"`

	Cluster       string `schema:"cluster,required"`
	Namespace     string `schema:"namespace,required"`
	ComponentType string `schema:"component_type,required"`
	ContainerName string `schema:"container_name"`

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

	tailLines := opt.TailLines
	if *tailLines == 0 {
		tailLines = nil
	}

	return &v1.PodLogOptions{
		Container:    opt.ContainerName,
		Follow:       opt.Follow,
		Previous:     opt.Previous,
		SinceSeconds: opt.SinceSeconds,
		SinceTime:    &sinceTime,
		Timestamps:   opt.Timestamps,
		TailLines:    tailLines,
		LimitBytes:   opt.LimitBytes,
	}
}
