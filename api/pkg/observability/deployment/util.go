package deployment

import (
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
)

const (
	timeoutReason         = "ProgressDeadlineExceeded"
	k8sRevisionAnnotation = "deployment.kubernetes.io/revision"
)

func getDeploymentCondition(status appsv1.DeploymentStatus, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

func getDeploymentRevision(depl *appsv1.Deployment) (int64, error) {
	v, ok := depl.GetAnnotations()[PublisherRevisionAnnotationKey]
	if !ok {
		return 0, nil
	}
	return strconv.ParseInt(v, 10, 64)
}
