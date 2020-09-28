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

package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

type Config struct {
	Address string
	Token   string
}

type ClusterSecret struct {
	Endpoint   string
	CaCert     string
	ClientCert string
	ClientKey  string
}

type VaultClient interface {
	GetClusterSecret(clusterName string) (*ClusterSecret, error)
}

type secretReader interface {
	Read(path string) (*api.Secret, error)
}

type vaultClient struct {
	secretReader
}

const (
	masterIPKey   = "master_ip"
	clientKeyKey  = "client_key"
	clientCertKey = "client_certificate"
	caCertKey     = "certs"
)

func NewVaultClient(cfg *Config) (VaultClient, error) {
	vaultCfg := &api.Config{
		Address: cfg.Address,
	}
	client, err := api.NewClient(vaultCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create vault client")
	}

	client.SetToken(cfg.Token)

	return newVaultClient(client.Logical())
}

func newVaultClient(secretReader secretReader) (VaultClient, error) {
	return &vaultClient{
		secretReader: secretReader,
	}, nil
}

func (v *vaultClient) GetClusterSecret(clusterName string) (*ClusterSecret, error) {
	secretPath := fmt.Sprintf("secret/%s", clusterName)
	secret, err := v.Read(secretPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read secret %s", secretPath)
	}

	if secret == nil {
		err := errors.New("secret is nil")
		return nil, errors.Wrapf(err, "unable to read secret %s", secretPath)
	}

	masterIP, err := getSecretAsString(secret, masterIPKey)
	if err != nil {
		return nil, err
	}

	clientKey, err := getSecretAsString(secret, clientKeyKey)
	if err != nil {
		return nil, err
	}

	clientCert, err := getSecretAsString(secret, clientCertKey)
	if err != nil {
		return nil, err
	}

	caCert, err := getSecretAsString(secret, caCertKey)
	if err != nil {
		return nil, err
	}

	return &ClusterSecret{
		Endpoint:   masterIP,
		ClientCert: clientCert,
		ClientKey:  clientKey,
		CaCert:     caCert,
	}, nil
}

func getSecretAsString(secret *api.Secret, key string) (string, error) {
	raw, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("unable to get %s", key)
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast %s as string", key)
	}

	return value, nil
}
