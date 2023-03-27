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
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/typed/core/v1/fake"
	fake2 "k8s.io/client-go/kubernetes/typed/rbac/v1/fake"
	ktesting "k8s.io/client-go/testing"

	jobspec "github.com/caraml-dev/merlin-pyspark-app/pkg/spec"
)

const (
	defaultNamespace = "spark-jobs"
)

func TestCreateJobSpecConfigMap(t *testing.T) {
	tests := []struct {
		name         string
		spec         *jobspec.PredictionJob
		want         *corev1.ConfigMap
		wantError    bool
		wantErrorMsg string
	}{
		{
			name: "nominal",
			spec: &jobspec.PredictionJob{
				Version: "v1",
				Kind:    "PredictionJob",
				Name:    jobName,
				Source: &jobspec.PredictionJob_BigquerySource{
					BigquerySource: &jobspec.BigQuerySource{
						Table: "project.dataset.table_iris",
						Features: []string{
							"sepal_length",
							"sepal_width",
							"petal_length",
							"petal_width",
						},
					},
				},
				Model: &jobspec.Model{
					Type: jobspec.ModelType_PYFUNC_V2,
					Uri:  "gs://bucket-name/e2e/artifacts/model",
					Result: &jobspec.Model_ModelResult{
						Type: jobspec.ResultType_INTEGER,
					},
				},
				Sink: &jobspec.PredictionJob_BigquerySink{
					BigquerySink: &jobspec.BigQuerySink{
						Table:         "project.dataset.table_iris_result",
						StagingBucket: "bucket-name",
						ResultColumn:  "prediction",
						SaveMode:      jobspec.SaveMode_OVERWRITE,
						Options: map[string]string{
							"project": "project-sample",
						},
					}},
			},
			want: &corev1.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Name:      jobName,
					Namespace: defaultNamespace,
				},
				Data: map[string]string{
					jobSpecFileName: `bigquerySink:
  options:
    project: project-sample
  resultColumn: prediction
  saveMode: OVERWRITE
  stagingBucket: bucket-name
  table: project.dataset.table_iris_result
bigquerySource:
  features:
  - sepal_length
  - sepal_width
  - petal_length
  - petal_width
  table: project.dataset.table_iris
kind: PredictionJob
model:
  result:
    type: INTEGER
  type: PYFUNC_V2
  uri: gs://bucket-name/e2e/artifacts/model
name: merlin-job
version: v1
`,
				},
			},
		},

		{
			name: "failed creation",
			spec: &jobspec.PredictionJob{
				Version: "v1",
				Kind:    "PredictionJob",
				Name:    jobName,
				Source: &jobspec.PredictionJob_BigquerySource{
					BigquerySource: &jobspec.BigQuerySource{
						Table: "project.dataset.table_iris",
						Features: []string{
							"sepal_length",
							"sepal_width",
							"petal_length",
							"petal_width",
						},
					},
				},
				Model: &jobspec.Model{
					Type: jobspec.ModelType_PYFUNC_V2,
					Uri:  "gs://bucket-name/e2e/artifacts/model",
					Result: &jobspec.Model_ModelResult{
						Type: jobspec.ResultType_INTEGER,
					},
				},
				Sink: &jobspec.PredictionJob_BigquerySink{
					BigquerySink: &jobspec.BigQuerySink{
						Table:         "project.dataset.table_iris_result",
						StagingBucket: "bucket-name",
						ResultColumn:  "prediction",
						SaveMode:      jobspec.SaveMode_OVERWRITE,
						Options: map[string]string{
							"project": "project-sample",
						},
					}},
			},
			wantError:    true,
			wantErrorMsg: "failed creating job specification config map",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClient := clientfake.NewSimpleClientset()
			fakeCm := fakeClient.CoreV1().ConfigMaps(defaultNamespace).(*fake.FakeConfigMaps)
			if test.wantError {
				fakeCm.Fake.PrependReactor("create", "configmaps",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("error creating config map")
					},
				)
			} else {
				fakeCm.Fake.PrependReactor("create", "configmaps",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						createAction := action.(ktesting.CreateAction)
						return true, createAction.GetObject(), nil
					},
				)
			}

			manifestManager := NewManifestManager(fakeClient)
			cmName, err := manifestManager.CreateJobSpec(context.Background(), test.spec.Name, defaultNamespace, test.spec)
			if test.wantError {
				assert.Error(t, err)
				assert.Equal(t, test.wantErrorMsg, err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.spec.Name, cmName)

			actions := fakeCm.Fake.Actions()
			assert.Len(t, actions, 1)
			createAction := actions[0].(ktesting.CreateAction)
			assert.Equal(t, test.want, createAction.GetObject().(*corev1.ConfigMap))
		})
	}
}

func TestDeleteJobSpecConfigMap(t *testing.T) {
	tests := []struct {
		name          string
		configMapName string
		wantError     bool
		wantErrorMsg  string
	}{
		{
			name:          "nominal",
			configMapName: jobName,
		},
		{
			name:          "error deleting",
			configMapName: jobName,
			wantError:     true,
			wantErrorMsg:  fmt.Sprintf("failed deleting configmap %s in namespace %s", jobName, defaultNamespace),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClient := clientfake.NewSimpleClientset()
			fakeCm := fakeClient.CoreV1().ConfigMaps(defaultNamespace).(*fake.FakeConfigMaps)
			if test.wantError {
				fakeCm.Fake.PrependReactor("delete", "configmaps",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("error deleting config map")
					},
				)
			} else {
				fakeCm.Fake.PrependReactor("delete", "configmaps",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, nil
					},
				)
			}

			manifestManager := NewManifestManager(fakeClient)
			err := manifestManager.DeleteJobSpec(context.Background(), test.configMapName, defaultNamespace)
			if test.wantError {
				assert.Error(t, err)
				assert.Equal(t, test.wantErrorMsg, err.Error())
				return
			}

			actions := fakeCm.Fake.Actions()
			assert.Len(t, actions, 1)
			deleteAction := actions[0].(ktesting.DeleteAction)
			assert.Equal(t, test.configMapName, deleteAction.GetName())
			assert.Equal(t, defaultNamespace, deleteAction.GetNamespace())
		})
	}
}

func TestCreateSecret(t *testing.T) {
	secret := "string secret"
	tests := []struct {
		name         string
		jobName      string
		data         string
		want         *corev1.Secret
		wantError    bool
		wantErrorMsg string
	}{
		{
			name:    "nominal",
			jobName: jobName,
			data:    secret,
			want: &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      jobName,
					Namespace: defaultNamespace,
				},
				StringData: map[string]string{
					serviceAccountFileName: secret,
				},
				Type: corev1.SecretTypeOpaque,
			},
		},
		{
			name:         "failed creation",
			jobName:      jobName,
			wantError:    true,
			wantErrorMsg: fmt.Sprintf("failed creating secret %s in namespace %s", jobName, defaultNamespace),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClient := clientfake.NewSimpleClientset()
			fakeSecretClient := fakeClient.CoreV1().Secrets(defaultNamespace).(*fake.FakeSecrets)
			if test.wantError {
				fakeSecretClient.Fake.PrependReactor("create", "secrets",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("error creating secret")
					},
				)
			} else {
				fakeSecretClient.Fake.PrependReactor("create", "secrets",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						createAction := action.(ktesting.CreateAction)
						return true, createAction.GetObject(), nil
					},
				)
			}

			manifestManager := NewManifestManager(fakeClient)
			secretName, err := manifestManager.CreateSecret(context.Background(), test.jobName, defaultNamespace, test.data)
			if test.wantError {
				assert.Error(t, err)
				assert.Equal(t, test.wantErrorMsg, err.Error())
				return
			}

			assert.Equal(t, test.jobName, secretName)
			actions := fakeSecretClient.Fake.Actions()
			assert.Len(t, actions, 1)
			createAction := actions[0].(ktesting.CreateAction)
			assert.Equal(t, test.want, createAction.GetObject())
		})
	}
}

func TestDeleteSecret(t *testing.T) {
	tests := []struct {
		name         string
		secretName   string
		wantError    bool
		wantErrorMsg string
	}{
		{
			name:       "nominal",
			secretName: jobName,
		},
		{
			name:         "error deleting",
			secretName:   jobName,
			wantError:    true,
			wantErrorMsg: fmt.Sprintf("failed deleting secret %s in namespace %s", jobName, defaultNamespace),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClient := clientfake.NewSimpleClientset()
			fakeSecrets := fakeClient.CoreV1().Secrets(defaultNamespace).(*fake.FakeSecrets)
			if test.wantError {
				fakeSecrets.Fake.PrependReactor("delete", "secrets",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("error deleting config map")
					},
				)
			} else {
				fakeSecrets.Fake.PrependReactor("delete", "secrets",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, nil
					},
				)
			}

			manifestManager := NewManifestManager(fakeClient)
			err := manifestManager.DeleteSecret(context.Background(), test.secretName, defaultNamespace)
			if test.wantError {
				assert.Error(t, err)
				assert.Equal(t, test.wantErrorMsg, err.Error())
				return
			}

			actions := fakeSecrets.Fake.Actions()
			assert.Len(t, actions, 1)
			deleteAction := actions[0].(ktesting.DeleteAction)
			assert.Equal(t, test.secretName, deleteAction.GetName())
			assert.Equal(t, defaultNamespace, deleteAction.GetNamespace())
		})
	}
}

func TestCreateDriverAuthorization(t *testing.T) {
	saName, roleName, roleBindingName := createAuthorizationResourceNames(defaultNamespace)
	tests := []struct {
		name               string
		namespace          string
		authResourceExists bool
		wantServiceAccount *corev1.ServiceAccount
		wantRole           *rbacv1.Role
		wantRoleBinding    *rbacv1.RoleBinding
	}{
		{
			name:      "nominal",
			namespace: defaultNamespace,
			wantServiceAccount: &corev1.ServiceAccount{
				ObjectMeta: v1.ObjectMeta{
					Name:      saName,
					Namespace: defaultNamespace,
				},
			},
			wantRole: &rbacv1.Role{
				ObjectMeta: v1.ObjectMeta{
					Name:      roleName,
					Namespace: defaultNamespace,
				},
				Rules: defaultSparkDriverRoleRules,
			},
			wantRoleBinding: &rbacv1.RoleBinding{
				ObjectMeta: v1.ObjectMeta{
					Name:      roleBindingName,
					Namespace: defaultNamespace,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Namespace: defaultNamespace,
						Name:      saName,
					},
				},
				RoleRef: rbacv1.RoleRef{
					Kind:     "Role",
					APIGroup: "rbac.authorization.k8s.io",
					Name:     roleName,
				},
			},
		},
		{
			name:               "authz-resource-exists",
			namespace:          defaultNamespace,
			authResourceExists: true,
			wantServiceAccount: &corev1.ServiceAccount{
				ObjectMeta: v1.ObjectMeta{
					Name:      saName,
					Namespace: defaultNamespace,
				},
			},
			wantRole: &rbacv1.Role{
				ObjectMeta: v1.ObjectMeta{
					Name:      roleName,
					Namespace: defaultNamespace,
				},
				Rules: defaultSparkDriverRoleRules,
			},
			wantRoleBinding: &rbacv1.RoleBinding{
				ObjectMeta: v1.ObjectMeta{
					Name:      roleBindingName,
					Namespace: defaultNamespace,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Namespace: defaultNamespace,
						Name:      saName,
					},
				},
				RoleRef: rbacv1.RoleRef{
					Kind:     "Role",
					APIGroup: "rbac.authorization.k8s.io",
					Name:     roleName,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClient := clientfake.NewSimpleClientset()
			fakeSaClient, fakeRoleClient, fakeRoleBindingClient := createAuthorizationFakes(fakeClient, test.namespace)
			mockDriverAuthorizationCreation(fakeSaClient, fakeRoleClient, fakeRoleBindingClient, test.authResourceExists)

			manifestManager := NewManifestManager(fakeClient)
			_, err := manifestManager.CreateDriverAuthorization(context.Background(), test.namespace)
			assert.NoError(t, err)

			expNumberOfAction := 3
			if !test.authResourceExists {
				expNumberOfAction += 3
			}
			assert.Len(t, fakeClient.Fake.Actions(), expNumberOfAction)

			if !test.authResourceExists {
				getSaAction := fakeClient.Fake.Actions()[0].(ktesting.GetAction)
				assert.Equal(t, test.wantServiceAccount.Name, getSaAction.GetName())
				assert.Equal(t, test.wantServiceAccount.Namespace, getSaAction.GetNamespace())

				getRoleAction := fakeClient.Fake.Actions()[2].(ktesting.GetAction)
				assert.Equal(t, test.wantRole.Name, getRoleAction.GetName())
				assert.Equal(t, test.wantRole.Namespace, getRoleAction.GetNamespace())

				getRoleBindingAction := fakeClient.Fake.Actions()[4].(ktesting.GetAction)
				assert.Equal(t, test.wantRoleBinding.Name, getRoleBindingAction.GetName())
				assert.Equal(t, test.wantRoleBinding.Namespace, getRoleBindingAction.GetNamespace())

				createSaAction := fakeClient.Fake.Actions()[1].(ktesting.CreateAction)
				assert.Equal(t, test.wantServiceAccount, createSaAction.GetObject())

				createRoleAction := fakeClient.Fake.Actions()[3].(ktesting.CreateAction)
				assert.Equal(t, test.wantRole, createRoleAction.GetObject())

				createRoleBindingAction := fakeClient.Fake.Actions()[5].(ktesting.CreateAction)
				assert.Equal(t, test.wantRoleBinding, createRoleBindingAction.GetObject())
			} else {
				getSaAction := fakeClient.Fake.Actions()[0].(ktesting.GetAction)
				assert.Equal(t, test.wantServiceAccount.Name, getSaAction.GetName())
				assert.Equal(t, test.wantServiceAccount.Namespace, getSaAction.GetNamespace())

				getRoleAction := fakeClient.Fake.Actions()[1].(ktesting.GetAction)
				assert.Equal(t, test.wantRole.Name, getRoleAction.GetName())
				assert.Equal(t, test.wantRole.Namespace, getRoleAction.GetNamespace())

				getRoleBindingAction := fakeClient.Fake.Actions()[2].(ktesting.GetAction)
				assert.Equal(t, test.wantRoleBinding.Name, getRoleBindingAction.GetName())
				assert.Equal(t, test.wantRoleBinding.Namespace, getRoleBindingAction.GetNamespace())
			}
		})
	}
}

func TestDeleteDriverAuthorization(t *testing.T) {
	tests := []struct {
		name         string
		namespace    string
		wantError    bool
		wantErrorMsg string
	}{
		{
			name:      "nominal",
			namespace: defaultNamespace,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClient := clientfake.NewSimpleClientset()
			fakeSaClient, fakeRoleClient, fakeRoleBindingClient := createAuthorizationFakes(fakeClient, test.namespace)
			mockDriverAuthorizationDeletion(fakeSaClient, fakeRoleClient, fakeRoleBindingClient)
			manifestManager := NewManifestManager(fakeClient)
			err := manifestManager.DeleteDriverAuthorization(context.Background(), test.namespace)
			if test.wantError {
				assert.Error(t, err)
				assert.Equal(t, test.wantErrorMsg, err.Error())
				return
			}

			assert.Len(t, fakeClient.Fake.Actions(), 3)

			saName, roleName, roleBindingName := createAuthorizationResourceNames(test.namespace)

			getSaAction := fakeClient.Fake.Actions()[2].(ktesting.DeleteAction)
			assert.Equal(t, saName, getSaAction.GetName())
			assert.Equal(t, test.namespace, getSaAction.GetNamespace())

			getRoleAction := fakeClient.Fake.Actions()[1].(ktesting.DeleteAction)
			assert.Equal(t, roleName, getRoleAction.GetName())
			assert.Equal(t, test.namespace, getRoleAction.GetNamespace())

			getRoleBindingAction := fakeClient.Fake.Actions()[0].(ktesting.DeleteAction)
			assert.Equal(t, roleBindingName, getRoleBindingAction.GetName())
			assert.Equal(t, test.namespace, getRoleBindingAction.GetNamespace())
		})
	}
}

func mockDriverAuthorizationDeletion(fakeSaClient *fake.FakeServiceAccounts, fakeRolesClient *fake2.FakeRoles, fakeRoleBindingsClient *fake2.FakeRoleBindings) {
	fakeSaClient.Fake.PrependReactor("delete", "serviceaccounts", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})
	fakeRolesClient.Fake.PrependReactor("delete", "roles", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})
	fakeRoleBindingsClient.Fake.PrependReactor("delete", "rolebindings", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})
}

func createAuthorizationFakes(client *clientfake.Clientset, namespace string) (fakeSaClient *fake.FakeServiceAccounts, fakeRolesClient *fake2.FakeRoles, fakeRoleBindingsClient *fake2.FakeRoleBindings) {
	fakeSaClient = client.CoreV1().ServiceAccounts(namespace).(*fake.FakeServiceAccounts)
	fakeRolesClient = client.RbacV1().Roles(namespace).(*fake2.FakeRoles)
	fakeRoleBindingsClient = client.RbacV1().RoleBindings(namespace).(*fake2.FakeRoleBindings)
	return
}

func mockDriverAuthorizationCreation(fakeSaClient *fake.FakeServiceAccounts, fakeRolesClient *fake2.FakeRoles, fakeRoleBindingsClient *fake2.FakeRoleBindings, authResourceExists bool) {
	if !authResourceExists {
		fakeSaClient.Fake.PrependReactor("get", "serviceaccounts", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, kerrors.NewNotFound(schema.ParseGroupResource("v1"), action.(ktesting.GetAction).GetName())
		})
		fakeRolesClient.Fake.PrependReactor("get", "roles", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, kerrors.NewNotFound(schema.ParseGroupResource("v1"), action.(ktesting.GetAction).GetName())
		})
		fakeRoleBindingsClient.Fake.PrependReactor("get", "rolebindings", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, kerrors.NewNotFound(schema.ParseGroupResource("v1"), action.(ktesting.GetAction).GetName())
		})

		fakeSaClient.Fake.PrependReactor("create", "serviceaccounts", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			createAction := action.(ktesting.CreateAction)
			return true, createAction.GetObject(), nil
		})
		fakeRolesClient.Fake.PrependReactor("create", "roles", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			createAction := action.(ktesting.CreateAction)
			return true, createAction.GetObject(), nil
		})
		fakeRoleBindingsClient.Fake.PrependReactor("create", "rolebindings", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			createAction := action.(ktesting.CreateAction)
			return true, createAction.GetObject(), nil
		})
	} else {
		fakeSaClient.Fake.PrependReactor("get", "serviceaccounts", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			getAction := action.(ktesting.GetAction)
			return true, &corev1.ServiceAccount{
				ObjectMeta: v1.ObjectMeta{
					Name:      getAction.GetName(),
					Namespace: getAction.GetNamespace(),
				},
			}, nil
		})
		fakeRolesClient.Fake.PrependReactor("get", "roles", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			getAction := action.(ktesting.GetAction)
			return true, &rbacv1.Role{
				ObjectMeta: v1.ObjectMeta{
					Name:      getAction.GetName(),
					Namespace: getAction.GetNamespace(),
				},
			}, nil
		})
		fakeRoleBindingsClient.Fake.PrependReactor("get", "rolebindings", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			getAction := action.(ktesting.GetAction)
			return true, &rbacv1.RoleBinding{
				ObjectMeta: v1.ObjectMeta{
					Name:      getAction.GetName(),
					Namespace: getAction.GetNamespace(),
				},
			}, nil
		})
	}
}
