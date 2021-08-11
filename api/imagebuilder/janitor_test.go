package imagebuilder

import (
	"testing"
	"time"

	"github.com/gojek/merlin/cluster/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJanitor_CleanJobs(t *testing.T) {
	namespace := "test-namespace"
	mockController := &mocks.Controller{}
	retention, _ := time.ParseDuration("1h")

	j := NewJanitor(mockController, JanitorConfig{BuildNamespace: namespace, Retention: retention})
	assert.NotNil(t, j)

	mockController.On("DeleteJobs", namespace, mock.Anything, mock.MatchedBy(func(listOptions metav1.ListOptions) bool {
		return listOptions.LabelSelector == labelOrchestratorName+"=merlin"
	})).Return(nil)

	j.CleanJobs()
}
