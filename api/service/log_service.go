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
	ImageBuilderLabelKey       = "job-name"
	KnativeServiceLabelKey     = "serving.knative.dev/service"
	BatchPredictionJobLabelKey = "prediction-job-id"

	prefixAddPod          = "pod"
	prefixAddContainer    = "container"
	prefixAddPodContainer = "pod_and_container"
)

var (
	skippedContainers = []string{"queue-proxy"}
)

// LogLine represents a single log line from a container running in Kubernetes.
type LogLine struct {
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

func (l LogLine) GenerateText(options LogQuery) string {
	text := []string{
		l.Timestamp.Format(time.RFC3339),
	}
	if options.Prefix == prefixAddPodContainer {
		text = append(text, l.PodName+"/"+l.ContainerName)
	} else if options.Prefix == prefixAddPod {
		text = append(text, l.PodName)
	} else if options.Prefix == prefixAddContainer {
		text = append(text, l.ContainerName)
	}
	text = append(text, l.TextPayload)
	return strings.Join(text, " ") + "\n"
}

type ReadLogStream struct {
	Stream        io.ReadCloser
	PodName       string
	ContainerName string
}

type LogService interface {
	StreamLogs(logLines chan string, stopCh chan struct{}, options *LogQuery) error
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

func (l logService) StreamLogs(logLines chan string, stopCh chan struct{}, options *LogQuery) error {
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
			allLogLines := make([]LogLine, 0)

			for _, pod := range pods.Items {
				for _, initContainer := range pod.Spec.InitContainers {
					options.ContainerName = initContainer.Name

					logLines := l.getContainerLogs(clusterClient, namespace, pod.Name, options)
					allLogLines = append(allLogLines, logLines...)
				}

				for _, container := range pod.Spec.Containers {
					for _, skippedContainer := range skippedContainers {
						if container.Name == skippedContainer {
							break
						}

						options.ContainerName = container.Name

						logLines := l.getContainerLogs(clusterClient, namespace, pod.Name, options)
						allLogLines = append(allLogLines, logLines...)
					}
				}
			}

			// Sort all the logs by timestamp in ascending order
			sort.Slice(allLogLines, func(i, j int) bool {
				return allLogLines[i].Timestamp.Before(allLogLines[j].Timestamp)
			})

			for _, logLine := range allLogLines {
				logLines <- logLine.GenerateText(*options)
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
	case models.ImageBuilderComponentType:
		if query.PredictionJobID == "" {
			return ImageBuilderLabelKey + "=" + query.ProjectName + "-" + query.ModelName + "-" + query.VersionID
		} else {
			return ImageBuilderLabelKey + "=batch-" + query.ProjectName + "-" + query.ModelName + "-" + query.VersionID
		}
	case models.ModelComponentType:
		return KnativeServiceLabelKey + "=" + query.ModelName + "-" + query.VersionID + "-predictor-default"
	case models.TransformerComponentType:
		return KnativeServiceLabelKey + "=" + query.ModelName + "-" + query.VersionID + "-transformer-default"
	case models.BatchJobDriverComponentType:
		return "spark-role=driver," + BatchPredictionJobLabelKey + "=" + query.PredictionJobID
	case models.BatchJobExecutorComponentType:
		return "spark-role=executor," + BatchPredictionJobLabelKey + "=" + query.PredictionJobID
	}
	return ""
}

func (l logService) getContainerLogs(clusterClient corev1.CoreV1Interface, namespace, podName string, options *LogQuery) []LogLine {
	stream, err := clusterClient.Pods(namespace).GetLogs(podName, options.ToKubernetesLogOption()).Stream()
	if err != nil {
		// Error is handled here by logging it rather than returned because the caller usually does not know how to
		// handle it. Example of what can trigger ListLogLines error: while the container is being created/terminated
		// Kubernetes API server will return error when logs are requested. In such case, it is better to return
		// empty logs and let the caller retry after the container becomes ready eventually.
		log.Warnf("Failed to ListLogLines: %s", err.Error())
		return nil
	}

	scanner := bufio.NewScanner(stream)
	logLines := make([]LogLine, 0)
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

		log := LogLine{
			Timestamp:     timestamp,
			Namespace:     namespace,
			PodName:       podName,
			ContainerName: options.ContainerName,
			TextPayload:   logLine[timestampIndex+1:],
		}

		logLines = append(logLines, log)
	}

	return logLines
}

type LogQuery struct {
	ProjectName     string `schema:"project_name"`
	ModelID         string `schema:"model_id"`
	ModelName       string `schema:"model_name"`
	VersionID       string `schema:"version_id"`
	PredictionJobID string `schema:"prediction_job_id"`

	Cluster       string `schema:"cluster,required"`
	Namespace     string `schema:"namespace,required"`
	ComponentType string `schema:"component_type,required"`
	ContainerName string `schema:"container_name"`

	Prefix string `schema:"prefix"`

	// Used for Kubernetes v1.PodLogOptions
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
