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

package mlp

import (
	"context"
	"fmt"

	"github.com/antihax/optional"

	"github.com/caraml-dev/mlp/api/client"
)

// SecretAPI is interface to mlp-api's Secret API.
type SecretAPI interface {
	ListSecrets(ctx context.Context, projectID int32) (Secrets, error)
	CreateSecret(ctx context.Context, projectID int32, secret Secret) (Secret, error)
	UpdateSecret(ctx context.Context, projectID int32, secret Secret) (Secret, error)
	DeleteSecret(ctx context.Context, secretID, projectID int32) error

	GetSecretByID(ctx context.Context, secretID, projectID int32) (Secret, error)
	GetSecretByName(ctx context.Context, secretName string, projectID int32) (Secret, error)
}

// Secret is mlp-api's Secret.
type Secret client.Secret

// Secrets is a list of mlp-api's Secret.
type Secrets []client.Secret

func (c *apiClient) ListSecrets(ctx context.Context, projectID int32) (Secrets, error) {
	secrets, _, err := c.client.SecretApi.V1ProjectsProjectIdSecretsGet(ctx, projectID) // nolint: bodyclose
	if err != nil {
		return nil, fmt.Errorf("mlp-api_ListSecrets: %w", err)
	}
	return secrets, nil
}

func (c *apiClient) CreateSecret(ctx context.Context, projectID int32, secret Secret) (Secret, error) {
	newSecret, _, err := c.client.SecretApi.V1ProjectsProjectIdSecretsPost(ctx, projectID, client.Secret(secret)) // nolint: bodyclose
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_CreateSecret: %w", err)
	}
	return Secret(newSecret), nil
}

func (c *apiClient) UpdateSecret(ctx context.Context, projectID int32, secret Secret) (Secret, error) {
	opt := &client.SecretApiV1ProjectsProjectIdSecretsSecretIdPatchOpts{
		Body: optional.NewInterface(client.Secret(secret)),
	}

	newSecret, _, err := c.client.SecretApi.V1ProjectsProjectIdSecretsSecretIdPatch(ctx, projectID, secret.ID, opt) // nolint: bodyclose
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_UpdateSecret: %w", err)
	}

	return Secret(newSecret), nil
}

func (c *apiClient) DeleteSecret(ctx context.Context, secretID, projectID int32) error {
	_, err := c.client.SecretApi.V1ProjectsProjectIdSecretsSecretIdDelete(ctx, projectID, secretID) // nolint: bodyclose
	if err != nil {
		return fmt.Errorf("mlp-api_DeleteSecret: %w", err)
	}
	return nil
}

func (c *apiClient) GetSecretByID(ctx context.Context, secretID, projectID int32) (Secret, error) {
	secret, _, err := c.client.SecretApi.V1ProjectsProjectIdSecretsSecretIdGet(ctx, projectID, secretID) // nolint: bodyclose
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_GetSecretByID: %w", err)
	}

	return Secret(secret), nil
}

func (c *apiClient) GetSecretByName(ctx context.Context, secretName string, projectID int32) (Secret, error) {
	secrets, _, err := c.client.SecretApi.V1ProjectsProjectIdSecretsGet(ctx, projectID) // nolint: bodyclose
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_GetSecretByName: %w", err)
	}

	for _, secret := range secrets {
		if secret.Name == secretName {
			return Secret(secret), nil
		}
	}

	return Secret{}, fmt.Errorf("mlp-GetSecretByName: Secret %s in Project %d not found", secretName, projectID)
}
