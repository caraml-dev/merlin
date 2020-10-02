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

	"github.com/gojek/mlp/api/client"
	"github.com/gojek/mlp/api/util"
)

// SecretAPI is interface to mlp-api's Secret API.
type SecretAPI interface {
	ListSecrets(ctx context.Context, projectID int32) (Secrets, error)
	CreateSecret(ctx context.Context, projectID int32, secret Secret) (Secret, error)
	UpdateSecret(ctx context.Context, projectID int32, secret Secret) (Secret, error)
	DeleteSecret(ctx context.Context, secretID, projectID int32) error

	GetSecretByIDandProjectID(ctx context.Context, secretID, projectID int32) (Secret, error)
	GetSecretByNameAndProjectID(ctx context.Context, secretName string, projectID int32) (Secret, error)
	GetPlainSecretByIDandProjectID(ctx context.Context, secretID, projectID int32) (Secret, error)
	GetPlainSecretByNameAndProjectID(ctx context.Context, secretName string, projectID int32) (Secret, error)
}

// Secret is mlp-api's Secret.
type Secret client.Secret

// Decrypt returns a copy of secret with decrypted data.
func (s Secret) Decrypt(passphrase string) (Secret, error) {
	encryptedData := s.Data
	decryptedData, err := util.Decrypt(encryptedData, passphrase)
	if err != nil {
		return Secret{}, err
	}

	return Secret{
		Id:   s.Id,
		Name: s.Name,
		Data: decryptedData,
	}, nil
}

// Secrets is a list of mlp-api's Secret.
type Secrets []client.Secret

func (c *apiClient) ListSecrets(ctx context.Context, projectID int32) (Secrets, error) {
	secrets, _, err := c.client.SecretApi.ProjectsProjectIdSecretsGet(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("mlp-api_ListSecrets: %s", err)
	}
	return secrets, nil
}

func (c *apiClient) CreateSecret(ctx context.Context, projectID int32, secret Secret) (Secret, error) {
	newSecret, _, err := c.client.SecretApi.ProjectsProjectIdSecretsPost(ctx, projectID, client.Secret(secret))
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_CreateSecret: %s", err)
	}
	return Secret(newSecret), nil
}

func (c *apiClient) UpdateSecret(ctx context.Context, projectID int32, secret Secret) (Secret, error) {
	opt := &client.SecretApiProjectsProjectIdSecretsSecretIdPatchOpts{
		Body: optional.NewInterface(client.Secret(secret)),
	}

	newSecret, _, err := c.client.SecretApi.ProjectsProjectIdSecretsSecretIdPatch(ctx, projectID, secret.Id, opt)
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_UpdateSecret: %s", err)
	}

	return Secret(newSecret), nil
}

func (c *apiClient) DeleteSecret(ctx context.Context, secretID, projectID int32) error {
	_, err := c.client.SecretApi.ProjectsProjectIdSecretsSecretIdDelete(ctx, projectID, secretID)
	if err != nil {
		return fmt.Errorf("mlp-api_DeleteSecret: %s", err)
	}
	return nil
}

func (c *apiClient) GetSecretByIDandProjectID(ctx context.Context, secretID, projectID int32) (Secret, error) {
	secrets, _, err := c.client.SecretApi.ProjectsProjectIdSecretsGet(ctx, projectID)
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_GetSecretByIDandProjectID: %s", err)
	}

	for _, secret := range secrets {
		if secret.Id == secretID {
			return Secret(secret), nil
		}
	}

	return Secret{}, fmt.Errorf("mlp-api_GetSecretByIDandProjectID: Secret %d in Project %d not found", secretID, projectID)
}

func (c *apiClient) GetSecretByNameAndProjectID(ctx context.Context, secretName string, projectID int32) (Secret, error) {
	secrets, _, err := c.client.SecretApi.ProjectsProjectIdSecretsGet(ctx, projectID)
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_GetSecretByNameAndProjectID: %s", err)
	}

	for _, secret := range secrets {
		if secret.Name == secretName {
			return Secret(secret), nil
		}
	}

	return Secret{}, fmt.Errorf("mlp-api_GetSecretByNameAndProjectID: Secret %s in Project %d not found", secretName, projectID)
}

func (c *apiClient) GetPlainSecretByIDandProjectID(ctx context.Context, secretID, projectID int32) (Secret, error) {
	secret, err := c.GetSecretByIDandProjectID(ctx, secretID, projectID)
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_GetPlainSecretByIDandProjectID: %s", err)
	}

	sec, err := secret.Decrypt(c.passphrase)
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_GetPlainSecretByIDandProjectID: error when decrypt secret data with id %d: %s", secretID, err)
	}

	return Secret(sec), nil
}

func (c *apiClient) GetPlainSecretByNameAndProjectID(ctx context.Context, secretName string, projectID int32) (Secret, error) {
	secret, err := c.GetSecretByNameAndProjectID(ctx, secretName, projectID)
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_GetPlainSecretByNameAndProjectID: %s", err)
	}

	sec, err := secret.Decrypt(c.passphrase)
	if err != nil {
		return Secret{}, fmt.Errorf("mlp-api_GetPlainSecretByNameAndProjectID: error when decrypt secret data with name %s: %s", secretName, err)
	}

	return Secret(sec), nil
}
