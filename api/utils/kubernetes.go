package utils

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

// ParsePodContainerStatuses parses the container statuses of a pod and returns a table of the statuses, a message, and a reason.
// If the container is running, it returns nothing.
func ParsePodContainerStatuses(containerStatuses []v1.ContainerStatus) (podTable, message, reason string) {
	podConditionHeaders := []string{"CONTAINER NAME", "STATUS", "EXIT CODE", "REASON"}
	podConditionRows := [][]string{}

	for _, status := range containerStatuses {
		if status.State.Waiting != nil {
			podConditionRows = append(podConditionRows, []string{
				status.Name,
				"Waiting",
				"-",
				status.State.Waiting.Reason,
			})
		}

		if status.State.Terminated != nil {
			podConditionRows = append(podConditionRows, []string{
				status.Name,
				"Terminated",
				fmt.Sprintf("%d", status.State.Terminated.ExitCode),
				status.State.Terminated.Reason,
			})

			message = status.State.Terminated.Message
			reason = status.State.Terminated.Reason
		}
	}

	// If there's no rows, this means the container is running, so we don't need to return anything
	if len(podConditionRows) == 0 {
		return
	}

	podTable = LogTable(podConditionHeaders, podConditionRows)
	return
}
