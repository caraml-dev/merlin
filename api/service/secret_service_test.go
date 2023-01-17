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
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/gojek/mlp/api/client"

	"github.com/gojek/merlin/mlp"
	mlpMock "github.com/gojek/merlin/mlp/mocks"
)

var (
	secret1 = mlp.Secret{
		ID:   int32(1),
		Name: "secret-1",
		Data: "data-1",
	}
	secret2 = mlp.Secret{
		ID:   int32(2),
		Name: "secret-2",
		Data: "data-2",
	}
)

func Test_secretService_List(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectID int32
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(*mlpMock.APIClient)
		want     mlp.Secrets
		wantErr  bool
	}{
		{
			name: "should success - return empty",
			args: args{
				ctx:       context.Background(),
				projectID: int32(1),
			},
			mockFunc: func(m *mlpMock.APIClient) {
				m.On("ListSecrets", mock.Anything, int32(1)).
					Return(mlp.Secrets{}, nil)
			},
			want:    mlp.Secrets{},
			wantErr: false,
		},
		{
			name: "should success - return secrets",
			args: args{
				ctx:       context.Background(),
				projectID: int32(1),
			},
			mockFunc: func(m *mlpMock.APIClient) {
				m.On("ListSecrets", mock.Anything, int32(1)).
					Return(mlp.Secrets{client.Secret(secret1)}, nil)
			},
			want:    mlp.Secrets{client.Secret(secret1)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMlpAPIClient := &mlpMock.APIClient{}
			tt.mockFunc(mockMlpAPIClient)

			ss := NewSecretService(mockMlpAPIClient)

			got, err := ss.List(tt.args.ctx, tt.args.projectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("secretService.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("secretService.List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_secretService_GetByIDandProjectID(t *testing.T) {
	type args struct {
		ctx       context.Context
		secretID  int32
		projectID int32
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(*mlpMock.APIClient)
		want     mlp.Secret
		wantErr  bool
	}{
		{
			name: "should success",
			args: args{
				ctx:       context.Background(),
				secretID:  int32(1),
				projectID: int32(1),
			},
			mockFunc: func(m *mlpMock.APIClient) {
				m.On("GetSecretByIDandProjectID", mock.Anything, int32(1), int32(1)).
					Return(secret1, nil)
			},
			want:    secret1,
			wantErr: false,
		},
		{
			name: "should return empty and no error if record not found",
			args: args{
				ctx:       context.Background(),
				secretID:  int32(1),
				projectID: int32(1),
			},
			mockFunc: func(m *mlpMock.APIClient) {
				m.On("GetSecretByIDandProjectID", mock.Anything, int32(1), int32(1)).
					Return(mlp.Secret{}, nil)
			},
			want:    mlp.Secret{},
			wantErr: false,
		},
		{
			name: "should return error if something going wrong when calling mlp-api",
			args: args{
				ctx:       context.Background(),
				secretID:  int32(1),
				projectID: int32(1),
			},
			mockFunc: func(m *mlpMock.APIClient) {
				m.On("GetSecretByIDandProjectID", mock.Anything, int32(1), int32(1)).
					Return(mlp.Secret{}, fmt.Errorf("internal server error"))
			},
			want:    mlp.Secret{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMlpAPIClient := &mlpMock.APIClient{}
			tt.mockFunc(mockMlpAPIClient)

			ss := NewSecretService(mockMlpAPIClient)

			got, err := ss.GetByIDandProjectID(tt.args.ctx, tt.args.secretID, tt.args.projectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("secretService.GetByIDandProjectID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("secretService.GetByIDandProjectID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_secretService_Create_Update(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectID int32
		secret    mlp.Secret
	}
	tests := []struct {
		name        string
		args        args
		mockFunc    func(*mlpMock.APIClient)
		wantCreated mlp.Secret
		wantUpdated mlp.Secret
		wantErr     bool
	}{
		{
			name: "should success",
			args: args{
				ctx:       context.Background(),
				projectID: int32(1),
				secret:    secret1,
			},
			mockFunc: func(m *mlpMock.APIClient) {
				m.On("CreateSecret", mock.Anything, int32(1), secret1).
					Return(secret1, nil)
				m.On("UpdateSecret", mock.Anything, int32(1), secret1).
					Return(secret2, nil)
			},
			wantCreated: secret1,
			wantUpdated: secret2,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMlpAPIClient := &mlpMock.APIClient{}
			tt.mockFunc(mockMlpAPIClient)

			ss := NewSecretService(mockMlpAPIClient)

			got, err := ss.Create(tt.args.ctx, tt.args.projectID, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("secretService.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantCreated) {
				t.Errorf("secretService.Create() = %v, want %v", got, tt.wantCreated)
			}

			got, err = ss.Update(tt.args.ctx, tt.args.projectID, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("secretService.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantUpdated) {
				t.Errorf("secretService.Update() = %v, want %v", got, tt.wantUpdated)
			}
		})
	}
}

func Test_secretService_Delete(t *testing.T) {
	type args struct {
		ctx       context.Context
		secretID  int32
		projectID int32
	}
	tests := []struct {
		name     string
		mockFunc func(*mlpMock.APIClient)
		args     args
		wantErr  bool
	}{
		{
			name: "should success",
			mockFunc: func(m *mlpMock.APIClient) {
				m.On("DeleteSecret", mock.Anything, int32(1), int32(1)).
					Return(nil)
			},
			args: args{
				ctx:       context.Background(),
				secretID:  int32(1),
				projectID: int32(1),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMlpAPIClient := &mlpMock.APIClient{}
			tt.mockFunc(mockMlpAPIClient)

			ss := NewSecretService(mockMlpAPIClient)

			if err := ss.Delete(tt.args.ctx, tt.args.secretID, tt.args.projectID); (err != nil) != tt.wantErr {
				t.Errorf("secretService.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
