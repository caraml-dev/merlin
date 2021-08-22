package imagebuilder

import (
	"reflect"
	"testing"
	"time"

	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/cluster/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	namespace    = "test-namespace"
	retention, _ = time.ParseDuration("1h")

	now       = metav1.NewTime(time.Now())
	yesterday = metav1.NewTime(time.Now().AddDate(0, 0, -1))

	activeJob1 = batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "batch-image-builder-not-expired-1",
			Labels: map[string]string{
				"gojek.com/orchestrator": "merlin",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "pyfunc-image-builder"},
					},
				},
			},
		},
		Status: batchv1.JobStatus{
			Active:    1,
			StartTime: &now,
		},
	}

	failedJob1 = batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "batch-image-builder-not-expired-1",
			Labels: map[string]string{
				"gojek.com/orchestrator": "merlin",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "pyfunc-image-builder"},
					},
				},
			},
		},
		Status: batchv1.JobStatus{
			Failed:    1,
			StartTime: &now,
		},
	}

	completedJob1 = batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "batch-image-builder-expired-1",
			Labels: map[string]string{
				"gojek.com/orchestrator": "merlin",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "pyfunc-image-builder"},
					},
				},
			},
		},
		Status: batchv1.JobStatus{
			Succeeded:      1,
			CompletionTime: &yesterday,
			StartTime:      &yesterday,
		},
	}

	completedJob2 = batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "batch-image-builder-expired-2",
			Labels: map[string]string{
				"gojek.com/orchestrator": "merlin",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "pyfunc-image-builder"},
					},
				},
			},
		},
		Status: batchv1.JobStatus{
			Succeeded:      1,
			CompletionTime: &yesterday,
			StartTime:      &yesterday,
		},
	}
)

func TestJanitor_CleanJobs(t *testing.T) {
	mc := &mocks.Controller{}
	j := NewJanitor(mc, JanitorConfig{BuildNamespace: namespace, Retention: retention})

	totalDelete := 0

	mc.On("ListJobs", namespace, labelOrchestratorName+"=merlin").
		Return(&batchv1.JobList{Items: []batchv1.Job{completedJob1, completedJob2, activeJob1, failedJob1}}, nil)

	mc.On("DeleteJob", namespace, completedJob1.Name, mock.Anything).
		Run(func(args mock.Arguments) {
			totalDelete++
		}).
		Return(nil)
	mc.On("DeleteJob", namespace, completedJob2.Name, mock.Anything).
		Run(func(args mock.Arguments) {
			totalDelete++
		}).
		Return(nil)

	j.CleanJobs()
	assert.Equal(t, 2, totalDelete)
}

func TestJanitor_getExpiredJobs(t *testing.T) {
	type fields struct {
		cc  cluster.Controller
		cfg JanitorConfig
	}
	tests := []struct {
		name    string
		fields  fields
		mockFn  func(mc *mocks.Controller)
		want    []batchv1.Job
		wantErr bool
	}{
		{
			name: "ListJobs returns empty",
			fields: fields{
				cfg: JanitorConfig{BuildNamespace: namespace, Retention: retention},
			},
			mockFn: func(mc *mocks.Controller) {
				mc.On("ListJobs", namespace, labelOrchestratorName+"=merlin").
					Return(&batchv1.JobList{}, nil)
			},
			want:    []batchv1.Job{},
			wantErr: false,
		},
		{
			name: "ListJobs returns one expired job",
			fields: fields{
				cfg: JanitorConfig{BuildNamespace: namespace, Retention: retention},
			},
			mockFn: func(mc *mocks.Controller) {
				mc.On("ListJobs", namespace, labelOrchestratorName+"=merlin").
					Return(&batchv1.JobList{Items: []batchv1.Job{completedJob1}}, nil)
			},
			want:    []batchv1.Job{completedJob1},
			wantErr: false,
		},
		{
			name: "ListJobs returns one expired job and one active job",
			fields: fields{
				cfg: JanitorConfig{BuildNamespace: namespace, Retention: retention},
			},
			mockFn: func(mc *mocks.Controller) {
				mc.On("ListJobs", namespace, labelOrchestratorName+"=merlin").
					Return(&batchv1.JobList{Items: []batchv1.Job{completedJob1, activeJob1}}, nil)
			},
			want:    []batchv1.Job{completedJob1},
			wantErr: false,
		},
		{
			name: "ListJobs returns one expired job and one active job and one failed job",
			fields: fields{
				cfg: JanitorConfig{BuildNamespace: namespace, Retention: retention},
			},
			mockFn: func(mc *mocks.Controller) {
				mc.On("ListJobs", namespace, labelOrchestratorName+"=merlin").
					Return(&batchv1.JobList{Items: []batchv1.Job{completedJob1, activeJob1, failedJob1}}, nil)
			},
			want:    []batchv1.Job{completedJob1},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &mocks.Controller{}

			j := NewJanitor(mc, tt.fields.cfg)

			tt.mockFn(mc)

			got, err := j.getExpiredJobs()
			if (err != nil) != tt.wantErr {
				t.Errorf("Janitor.getExpiredJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Janitor.getExpiredJobs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJanitor_deleteJobs_one(t *testing.T) {
	mc := &mocks.Controller{}
	j := NewJanitor(mc, JanitorConfig{BuildNamespace: namespace, Retention: retention})

	mc.On("DeleteJob", namespace, completedJob1.Name, mock.Anything).Return(nil)

	err := j.deleteJobs([]batchv1.Job{completedJob1})
	assert.Nil(t, err)
}

func TestJanitor_deleteJobs_two(t *testing.T) {
	mc := &mocks.Controller{}
	j := NewJanitor(mc, JanitorConfig{BuildNamespace: namespace, Retention: retention})

	totalDelete := 0

	mc.On("DeleteJob", namespace, completedJob1.Name, mock.Anything).
		Run(func(args mock.Arguments) {
			totalDelete++
		}).
		Return(nil)
	mc.On("DeleteJob", namespace, completedJob2.Name, mock.Anything).
		Run(func(args mock.Arguments) {
			totalDelete++
		}).
		Return(nil)

	err := j.deleteJobs([]batchv1.Job{completedJob1, completedJob2})
	assert.Nil(t, err)

	assert.Equal(t, 2, totalDelete)
}
