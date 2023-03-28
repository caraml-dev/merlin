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

package service

import (
	"context"

	"github.com/caraml-dev/merlin/mlp"
)

type SecretService interface {
	List(ctx context.Context, projectID int32) (mlp.Secrets, error)
	GetByIDandProjectID(ctx context.Context, secretID, projectID int32) (mlp.Secret, error)
	Create(ctx context.Context, projectID int32, secret mlp.Secret) (mlp.Secret, error)
	Update(ctx context.Context, projectID int32, secret mlp.Secret) (mlp.Secret, error)
	Delete(ctx context.Context, secretID, projectID int32) error
}

func NewSecretService(mlpAPIClient mlp.APIClient) SecretService {
	return &secretService{
		mlpAPIClient: mlpAPIClient,
	}
}

type secretService struct {
	mlpAPIClient mlp.APIClient
}

func (ss *secretService) List(ctx context.Context, projectID int32) (mlp.Secrets, error) {
	return ss.mlpAPIClient.ListSecrets(ctx, projectID)
}

func (ss *secretService) GetByIDandProjectID(ctx context.Context, secretID, projectID int32) (mlp.Secret, error) {
	return ss.mlpAPIClient.GetSecretByIDandProjectID(ctx, secretID, projectID)
}

func (ss *secretService) Create(ctx context.Context, projectID int32, secret mlp.Secret) (mlp.Secret, error) {
	return ss.mlpAPIClient.CreateSecret(ctx, projectID, secret)
}

func (ss *secretService) Update(ctx context.Context, projectID int32, secret mlp.Secret) (mlp.Secret, error) {
	return ss.mlpAPIClient.UpdateSecret(ctx, projectID, secret)
}

func (ss *secretService) Delete(ctx context.Context, secretID, projectID int32) error {
	return ss.mlpAPIClient.DeleteSecret(ctx, secretID, projectID)
}
