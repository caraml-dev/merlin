package utils

import (
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
)

func ParsePodContainerStatuses(containerStatuses []v1.ContainerStatus) (string, string, error) {
	var err error

	podConditionHeaders := []string{"CONTAINER NAME", "STATUS", "EXIT CODE", "REASON"}
	podConditionRows := [][]string{}
	message := ""

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

			err = errors.New(status.State.Terminated.Reason)
		}
	}

	// If there's no rows, this means the container is running, so we don't need to return anything
	if len(podConditionRows) == 0 {
		return "", "", nil
	}

	podTable := LogTable(podConditionHeaders, podConditionRows)
	return podTable, message, err
}
