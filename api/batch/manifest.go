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

package batch

import (
	"bytes"
	"context"
	"fmt"

	"github.com/caraml-dev/merlin-pyspark-app/pkg/spec"
	"github.com/caraml-dev/merlin/log"
	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ManifestManager interface {
	CreateJobSpec(ctx context.Context, predictionJobName string, namespace string, spec *spec.PredictionJob) (string, error)
	DeleteJobSpec(ctx context.Context, predictionJobName string, namespace string) error

	CreateSecret(ctx context.Context, predictionJobName string, namespace string, secretMap map[string]string) (string, error)
	DeleteSecret(ctx context.Context, predictionJobName string, namespace string) error

	CreateDriverAuthorization(ctx context.Context, namespace string) (string, error)
	DeleteDriverAuthorization(ctx context.Context, namespace string) error
}

var (
	jsonMarshaller              = &jsonpb.Marshaler{}
	defaultSparkDriverRoleRules = []v1.PolicyRule{
		{
			// Allow driver to manage pods
			APIGroups: []string{
				"", // indicates the core API group
			},
			Resources: []string{
				"pods",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			// Allow driver to manage services
			APIGroups: []string{
				"", // indicates the core API group
			},
			Resources: []string{
				"services",
			},
			Verbs: []string{
				"*",
			},
		},
		{
			// Allow driver to manage configmaps
			APIGroups: []string{
				"", // indicates the core API group
			},
			Resources: []string{
				"configmaps",
			},
			Verbs: []string{
				"*",
			},
		},
	}
)

type manifestManager struct {
	kubeClient kubernetes.Interface
}

func NewManifestManager(kubeClient kubernetes.Interface) ManifestManager {
	return &manifestManager{kubeClient: kubeClient}
}

func (m *manifestManager) CreateJobSpec(ctx context.Context, predictionJobName string, namespace string, spec *spec.PredictionJob) (string, error) {
	configYaml, err := toYamlString(spec)
	if err != nil {
		log.Errorf("failed converting prediction job spec to yaml: %v", err)
		return "", errors.New("failed converting prediction job spec to yaml")
	}

	cm, err := m.kubeClient.CoreV1().ConfigMaps(namespace).Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      predictionJobName,
			Namespace: namespace,
		},
		Data: map[string]string{
			jobSpecFileName: configYaml,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		log.Errorf("failed creating job specification config map %s in namespace %s: %v", predictionJobName, namespace, err)
		return "", errors.New("failed creating job specification config map")
	}

	return cm.Name, nil
}

func (m *manifestManager) DeleteJobSpec(ctx context.Context, predictionJobName string, namespace string) error {
	err := m.kubeClient.CoreV1().ConfigMaps(namespace).Delete(ctx, predictionJobName, metav1.DeleteOptions{})
	if client.IgnoreNotFound(err) != nil {
		log.Errorf("failed deleting configmap %s in namespace %s: %v", predictionJobName, namespace, err)
		return errors.Errorf("failed deleting configmap %s in namespace %s", predictionJobName, namespace)
	}
	return nil
}

func (m *manifestManager) CreateDriverAuthorization(ctx context.Context, namespace string) (string, error) {
	serviceAccountName, driverRoleName, driverRoleBindingName := createAuthorizationResourceNames(namespace)
	// create service account
	sa, err := m.kubeClient.CoreV1().ServiceAccounts(namespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return "", errors.Errorf("failed getting status of driver service account %s in namespace %s", serviceAccountName, namespace)
		}

		sa, err = m.kubeClient.CoreV1().ServiceAccounts(namespace).Create(ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		}, metav1.CreateOptions{})

		if err != nil {
			return "", errors.Errorf("failed creating driver service account %s in namespace %s", serviceAccountName, namespace)
		}
	}

	// create role
	role, err := m.kubeClient.RbacV1().Roles(namespace).Get(ctx, driverRoleName, metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return "", errors.Errorf("failed getting status of driver role %s in namespace %s", driverRoleName, namespace)
		}

		role, err = m.kubeClient.RbacV1().Roles(namespace).Create(ctx, &v1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      driverRoleName,
				Namespace: namespace,
			},
			Rules: defaultSparkDriverRoleRules,
		}, metav1.CreateOptions{})

		if err != nil {
			return "", errors.Errorf("failed creating driver roles %s in namespace %s", driverRoleName, namespace)
		}
	}

	// create role binding
	_, err = m.kubeClient.RbacV1().RoleBindings(namespace).Get(ctx, driverRoleBindingName, metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return "", errors.Errorf("failed getting status of driver rolebinding %s in namespace %s", driverRoleBindingName, namespace)
		}

		_, err = m.kubeClient.RbacV1().RoleBindings(namespace).Create(ctx, &v1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      driverRoleBindingName,
				Namespace: namespace,
			},
			Subjects: []v1.Subject{
				{
					Kind:      "ServiceAccount",
					Namespace: namespace,
					Name:      sa.Name,
				},
			},
			RoleRef: v1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     role.Name,
			},
		}, metav1.CreateOptions{})

		if err != nil {
			return "", errors.Errorf("failed creating driver roles binding %s in namespace %s", driverRoleBindingName, namespace)
		}
	}

	return sa.Name, nil
}

func (m *manifestManager) DeleteDriverAuthorization(ctx context.Context, namespace string) error {
	serviceAccountName, driverRoleName, driverRoleBindingName := createAuthorizationResourceNames(namespace)
	err := m.kubeClient.RbacV1().RoleBindings(namespace).Delete(ctx, driverRoleBindingName, metav1.DeleteOptions{})
	if client.IgnoreNotFound(err) != nil {
		return errors.Errorf("failed deleting driver roles binding %s in namespace %s", driverRoleBindingName, namespace)
	}
	err = m.kubeClient.RbacV1().Roles(namespace).Delete(ctx, driverRoleName, metav1.DeleteOptions{})
	if client.IgnoreNotFound(err) != nil {
		return errors.Errorf("failed deleting driver roles %s in namespace %s", driverRoleName, namespace)
	}
	err = m.kubeClient.CoreV1().ServiceAccounts(namespace).Delete(ctx, serviceAccountName, metav1.DeleteOptions{})
	if client.IgnoreNotFound(err) != nil {
		return errors.Errorf("failed deleting service account %s in namespace %s", serviceAccountName, namespace)
	}
	return nil
}

func (m *manifestManager) CreateSecret(ctx context.Context, predictionJobName string, namespace string, secretMap map[string]string) (string, error) {
	secret, err := m.kubeClient.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      predictionJobName,
			Namespace: namespace,
		},
		StringData: secretMap,
		Type:       corev1.SecretTypeOpaque,
	}, metav1.CreateOptions{})

	if err != nil {
		log.Errorf("failed creating secret %s in namespace %s: %v", predictionJobName, namespace, err)
		return "", errors.Errorf("failed creating secret %s in namespace %s", predictionJobName, namespace)
	}

	return secret.Name, nil
}

func (m *manifestManager) DeleteSecret(ctx context.Context, predictionJobName string, namespace string) error {
	err := m.kubeClient.CoreV1().Secrets(namespace).Delete(ctx, predictionJobName, metav1.DeleteOptions{})
	if client.IgnoreNotFound(err) != nil {
		log.Errorf("failed deleting secret %s in namespace %s: %v", predictionJobName, namespace, err)
		return errors.Errorf("failed deleting secret %s in namespace %s", predictionJobName, namespace)
	}
	return nil
}

func toYamlString(spec *spec.PredictionJob) (string, error) {
	buf := new(bytes.Buffer)
	err := jsonMarshaller.Marshal(buf, spec)
	if err != nil {
		return "", err
	}

	res, err := yaml.JSONToYAML(buf.Bytes())
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func createAuthorizationResourceNames(namespace string) (serviceAccountName, driverRoleName, driverRoleBindingName string) {
	serviceAccountName = fmt.Sprintf("%s-driver-sa", namespace)
	driverRoleName = fmt.Sprintf("%s-driver-role", namespace)
	driverRoleBindingName = fmt.Sprintf("%s-driver-role-binding", namespace)
	return
}
