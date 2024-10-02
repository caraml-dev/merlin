package deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_gcp_enrichIdentityToPod(t *testing.T) {
	singleContainerPodSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "worker",
				Image: "python:3.10",
				Command: []string{
					"python",
					"-m",
					"publisher",
					"+environment=config",
				},
				ImagePullPolicy: corev1.PullIfNotPresent,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "prom-metric",
						ContainerPort: 8000,
						Protocol:      corev1.ProtocolTCP,
					},
				},
			},
		},
	}

	multipleContainerPodSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "worker",
				Image: "python:3.10",
				Command: []string{
					"python",
					"-m",
					"publisher",
					"+environment=config",
				},
				ImagePullPolicy: corev1.PullIfNotPresent,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "env-secret-volume",
						MountPath: "/env",
						ReadOnly:  true,
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "prom-metric",
						ContainerPort: 8000,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "WORKERS",
						Value: "1",
					},
				},
			},
			{
				Name:  "sidecar",
				Image: "python:3.10",
				Command: []string{
					"./run.sh",
				},
				ImagePullPolicy: corev1.PullIfNotPresent,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "env-secret-volume",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: "secret-1",
					},
				},
			},
		},
	}
	type args struct {
		podSpec        corev1.PodSpec
		secretName     string
		containerNames []string
	}
	tests := []struct {
		name string
		args args
		want corev1.PodSpec
	}{
		{
			name: "update identity for one container",
			args: args{
				podSpec:        singleContainerPodSpec,
				secretName:     "service-account",
				containerNames: []string{"worker"},
			},
			want: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "worker",
						Image: "python:3.10",
						Command: []string{
							"python",
							"-m",
							"publisher",
							"+environment=config",
						},
						ImagePullPolicy: corev1.PullIfNotPresent,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          "prom-metric",
								ContainerPort: 8000,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "iam-secret",
								MountPath: "/iam/service-account",
								ReadOnly:  true,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "GOOGLE_APPLICATION_CREDENTIALS",
								Value: "/iam/service-account/service-account.json",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "iam-secret",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "service-account",
							},
						},
					},
				},
			},
		},
		{
			name: "update identity for one container, but no matching container name",
			args: args{
				podSpec:        singleContainerPodSpec,
				secretName:     "service-account",
				containerNames: []string{"not-exist"},
			},
			want: singleContainerPodSpec,
		},
		{
			name: "update identity for multiple containers",
			args: args{
				podSpec:        multipleContainerPodSpec,
				secretName:     "service-account",
				containerNames: []string{"worker", "sidecar"},
			},
			want: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "worker",
						Image: "python:3.10",
						Command: []string{
							"python",
							"-m",
							"publisher",
							"+environment=config",
						},
						ImagePullPolicy: corev1.PullIfNotPresent,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "env-secret-volume",
								MountPath: "/env",
								ReadOnly:  true,
							},
							{
								Name:      "iam-secret",
								MountPath: "/iam/service-account",
								ReadOnly:  true,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "WORKERS",
								Value: "1",
							},
							{
								Name:  "GOOGLE_APPLICATION_CREDENTIALS",
								Value: "/iam/service-account/service-account.json",
							},
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          "prom-metric",
								ContainerPort: 8000,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					},
					{
						Name:  "sidecar",
						Image: "python:3.10",
						Command: []string{
							"./run.sh",
						},
						ImagePullPolicy: corev1.PullIfNotPresent,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "iam-secret",
								MountPath: "/iam/service-account",
								ReadOnly:  true,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "GOOGLE_APPLICATION_CREDENTIALS",
								Value: "/iam/service-account/service-account.json",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "env-secret-volume",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "secret-1",
							},
						},
					},
					{
						Name: "iam-secret",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "service-account",
							},
						},
					},
				},
			},
		},
		{
			name: "update identity for multiple containers, only match one container",
			args: args{
				podSpec:        multipleContainerPodSpec,
				secretName:     "service-account",
				containerNames: []string{"worker"},
			},
			want: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "worker",
						Image: "python:3.10",
						Command: []string{
							"python",
							"-m",
							"publisher",
							"+environment=config",
						},
						ImagePullPolicy: corev1.PullIfNotPresent,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "env-secret-volume",
								MountPath: "/env",
								ReadOnly:  true,
							},
							{
								Name:      "iam-secret",
								MountPath: "/iam/service-account",
								ReadOnly:  true,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "WORKERS",
								Value: "1",
							},
							{
								Name:  "GOOGLE_APPLICATION_CREDENTIALS",
								Value: "/iam/service-account/service-account.json",
							},
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          "prom-metric",
								ContainerPort: 8000,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					},
					{
						Name:  "sidecar",
						Image: "python:3.10",
						Command: []string{
							"./run.sh",
						},
						ImagePullPolicy: corev1.PullIfNotPresent,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "env-secret-volume",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "secret-1",
							},
						},
					},
					{
						Name: "iam-secret",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "service-account",
							},
						},
					},
				},
			},
		},
		{
			name: "update identity for multiple containers, but none matching container name",
			args: args{
				podSpec:        multipleContainerPodSpec,
				secretName:     "service-account",
				containerNames: []string{"worker-1", "sidecar-1"},
			},
			want: multipleContainerPodSpec,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := enrichIdentityToPod(tt.args.podSpec, tt.args.secretName, tt.args.containerNames)
			assert.Equal(t, tt.want, got)
		})
	}
}
