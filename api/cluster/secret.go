package cluster

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/caraml-dev/merlin/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *controller) createSecret(ctx context.Context, secretName string, namespace string, secretMap map[string]string) (string, error) {
	_, err := c.clusterClient.Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	secretConfig := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: secretMap,
		Type:       corev1.SecretTypeOpaque,
	}

	if err == nil {
		secret, err := c.clusterClient.Secrets(namespace).Update(ctx, &secretConfig, metav1.UpdateOptions{})
		if err != nil {
			log.Errorf("failed updating secret %s in namespace %s: %v", secretName, namespace, err)
			return "", fmt.Errorf("failed updating secret %s in namespace %s: %w", secretName, namespace, err)
		}
		return secret.Name, nil
	}

	secret, err := c.clusterClient.Secrets(namespace).Create(ctx, &secretConfig, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("failed creating secret %s in namespace %s: %v", secretName, namespace, err)
		return "", fmt.Errorf("failed creating secret %s in namespace %s: %w", secretName, namespace, err)
	}
	return secret.Name, nil
}

func (c *controller) deleteSecret(ctx context.Context, secretName string, namespace string) error {
	err := c.clusterClient.Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("failed deleting secret %s in namespace %s: %w", secretName, namespace, err)
	}
	return nil
}
