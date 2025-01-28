package cluster

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/caraml-dev/merlin/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *controller) createSecret(ctx context.Context, secretName string, namespace string, secretMap map[string]string) (string, error) {
	secret, err := c.clusterClient.Secrets(namespace).Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: secretMap,
		Type:       corev1.SecretTypeOpaque,
	}, metav1.CreateOptions{})

	if err != nil {
		log.Errorf("failed creating secret %s in namespace %s: %v", secretName, namespace, err)
		return "", errors.Errorf("failed creating secret %s in namespace %s", secretName, namespace)
	}

	return secret.Name, nil
}

func (c *controller) deleteSecret(ctx context.Context, secretName string, namespace string) error {
	err := c.clusterClient.Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if client.IgnoreNotFound(err) != nil {
		return errors.Errorf("failed deleting secret %s in namespace %s", secretName, namespace)
	}
	return nil
}
