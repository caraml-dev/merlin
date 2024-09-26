package deployment

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func enrichIdentityToPod(podSpec corev1.PodSpec, secretName string, containerNames []string) corev1.PodSpec {
	secretVolume := createVolumeFromSecret(secretName)
	updatedPodSpec := podSpec.DeepCopy()

	containerExist := false
	containerNameLookup := make(map[string]bool)
	for _, containerName := range containerNames {
		containerNameLookup[containerName] = true
	}

	for idx, containerSpec := range updatedPodSpec.Containers {
		if val := containerNameLookup[containerSpec.Name]; !val {
			continue
		}

		containerExist = true
		mountPath := fmt.Sprintf("/iam/%s", secretName)
		volumeMount := corev1.VolumeMount{
			Name:      secretVolume.Name,
			MountPath: mountPath,
			ReadOnly:  true,
		}
		containerSpec.VolumeMounts = append(containerSpec.VolumeMounts, volumeMount)
		gcpCredentialEnvVar := corev1.EnvVar{
			Name:  "GOOGLE_APPLICATION_CREDENTIALS",
			Value: fmt.Sprintf("%s/service-account.json", mountPath),
		}
		containerSpec.Env = append(containerSpec.Env, gcpCredentialEnvVar)
		updatedPodSpec.Containers[idx] = containerSpec
	}

	if containerExist {
		updatedPodSpec.Volumes = append(updatedPodSpec.Volumes, secretVolume)
	}
	return *updatedPodSpec
}

func createVolumeFromSecret(secretName string) corev1.Volume {
	return corev1.Volume{
		Name: "iam-secret",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
}
