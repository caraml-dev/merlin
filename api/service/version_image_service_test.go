package service

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/imagebuilder/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_versionImageService_GetImage(t *testing.T) {
	type args struct {
		ctx     context.Context
		model   *models.Model
		version *models.Version
	}
	tests := []struct {
		name         string
		imageBuilder func() *mocks.ImageBuilder
		args         args
		want         models.VersionImage
		wantErr      bool
	}{
		{
			name: "success: image exists and job still exists",
			imageBuilder: func() *mocks.ImageBuilder {
				ib := &mocks.ImageBuilder{}
				ib.On("GetVersionImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(models.VersionImage{
						ProjectID: 1,
						ModelID:   1,
						VersionID: 1,
						ImageRef:  "ghcr.io/caraml-dev/project-model:1",
						Exists:    true,
					}, nil)
				ib.On("GetImageBuildingJobStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(models.ImageBuildingJobStatusSucceeded, nil)
				return ib
			},
			args: args{
				ctx: context.Background(),
				model: &models.Model{
					ID: 1,
					Project: mlp.Project{
						ID: 1,
					},
					Type: models.ModelTypePyFunc,
				},
				version: &models.Version{
					ID: 1,
				},
			},
			want: models.VersionImage{
				ProjectID: 1,
				ModelID:   1,
				VersionID: 1,
				ImageRef:  "ghcr.io/caraml-dev/project-model:1",
				Exists:    true,
				JobStatus: models.ImageBuildingJobStatusSucceeded,
			},
			wantErr: false,
		},
		{
			name: "success: image exists but job is not exists anymore",
			imageBuilder: func() *mocks.ImageBuilder {
				ib := &mocks.ImageBuilder{}
				ib.On("GetVersionImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(models.VersionImage{
						ProjectID: 1,
						ModelID:   1,
						VersionID: 1,
						ImageRef:  "ghcr.io/caraml-dev/project-model:1",
						Exists:    true,
					}, nil)
				ib.On("GetImageBuildingJobStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(models.ImageBuildingJobStatusUnknown, nil)
				return ib
			},
			args: args{
				ctx: context.Background(),
				model: &models.Model{
					ID: 1,
					Project: mlp.Project{
						ID: 1,
					},
					Type: models.ModelTypePyFunc,
				},
				version: &models.Version{
					ID: 1,
				},
			},
			want: models.VersionImage{
				ProjectID: 1,
				ModelID:   1,
				VersionID: 1,
				ImageRef:  "ghcr.io/caraml-dev/project-model:1",
				Exists:    true,
				JobStatus: models.ImageBuildingJobStatusUnknown,
			},
			wantErr: false,
		},
		{
			name: "success: image not exists but job still active",
			imageBuilder: func() *mocks.ImageBuilder {
				ib := &mocks.ImageBuilder{}
				ib.On("GetVersionImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(models.VersionImage{
						ProjectID: 1,
						ModelID:   1,
						VersionID: 1,
						ImageRef:  "ghcr.io/caraml-dev/project-model:1",
						Exists:    false,
					}, nil)
				ib.On("GetImageBuildingJobStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(models.ImageBuildingJobStatusActive, nil)
				return ib
			},
			args: args{
				ctx: context.Background(),
				model: &models.Model{
					ID: 1,
					Project: mlp.Project{
						ID: 1,
					},
					Type: models.ModelTypePyFunc,
				},
				version: &models.Version{
					ID: 1,
				},
			},
			want: models.VersionImage{
				ProjectID: 1,
				ModelID:   1,
				VersionID: 1,
				ImageRef:  "ghcr.io/caraml-dev/project-model:1",
				Exists:    false,
				JobStatus: models.ImageBuildingJobStatusActive,
			},
			wantErr: false,
		},
		{
			name: "invalid model type",
			imageBuilder: func() *mocks.ImageBuilder {
				ib := &mocks.ImageBuilder{}
				return ib
			},
			args: args{
				ctx: context.Background(),
				model: &models.Model{
					ID: 1,
					Project: mlp.Project{
						ID: 1,
					},
					Type: models.ModelTypeXgboost,
				},
				version: &models.Version{
					ID: 1,
				},
			},
			want:    models.VersionImage{},
			wantErr: true,
		},
		{
			name: "error getting image building job",
			imageBuilder: func() *mocks.ImageBuilder {
				ib := &mocks.ImageBuilder{}
				ib.On("GetVersionImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(models.VersionImage{
						ProjectID: 1,
						ModelID:   1,
						VersionID: 1,
						ImageRef:  "ghcr.io/caraml-dev/project-model:1",
						Exists:    true,
					}, nil)
				ib.On("GetImageBuildingJobStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(models.ImageBuildingJobStatusUnknown, fmt.Errorf("context deadline exceeded"))
				return ib
			},
			args: args{
				ctx: context.Background(),
				model: &models.Model{
					ID: 1,
					Project: mlp.Project{
						ID: 1,
					},
					Type: models.ModelTypePyFunc,
				},
				version: &models.Version{
					ID: 1,
				},
			},
			want: models.VersionImage{
				ProjectID: 1,
				ModelID:   1,
				VersionID: 1,
				ImageRef:  "ghcr.io/caraml-dev/project-model:1",
				Exists:    true,
				JobStatus: models.ImageBuildingJobStatusUnknown,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewVersionImageService(tt.imageBuilder(), tt.imageBuilder())

			got, err := s.GetImage(tt.args.ctx, tt.args.model, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("versionImageService.GetImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("versionImageService.GetImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_versionImageService_BuildImage(t *testing.T) {
	type args struct {
		ctx     context.Context
		model   *models.Model
		version *models.Version
		options *models.BuildImageOptions
	}
	tests := []struct {
		name         string
		imageBuilder func() *mocks.ImageBuilder
		args         args
		want         string
		wantErr      bool
	}{
		{
			name: "success",
			imageBuilder: func() *mocks.ImageBuilder {
				ib := &mocks.ImageBuilder{}
				ib.On("BuildImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("ghcr.io/caraml-dev/project-model:1", nil)
				return ib
			},
			args: args{
				ctx: context.Background(),
				model: &models.Model{
					ID: 1,
					Project: mlp.Project{
						ID: 1,
					},
					Type: models.ModelTypePyFuncV2,
				},
				version: &models.Version{
					ID: 1,
				},
				options: &models.BuildImageOptions{
					ResourceRequest: &models.ResourceRequest{
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
				},
			},
			want:    "ghcr.io/caraml-dev/project-model:1",
			wantErr: false,
		},
		{
			name: "invalid model type",
			imageBuilder: func() *mocks.ImageBuilder {
				ib := &mocks.ImageBuilder{}
				ib.On("BuildImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("ghcr.io/caraml-dev/project-model:1", nil)
				return ib
			},
			args: args{
				ctx: context.Background(),
				model: &models.Model{
					ID: 1,
					Project: mlp.Project{
						ID: 1,
					},
					Type: models.ModelTypeXgboost,
				},
				version: &models.Version{
					ID: 1,
				},
				options: &models.BuildImageOptions{
					ResourceRequest: &models.ResourceRequest{
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "failed building image",
			imageBuilder: func() *mocks.ImageBuilder {
				ib := &mocks.ImageBuilder{}
				ib.On("BuildImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("", fmt.Errorf("error building image: timeout"))
				return ib
			},
			args: args{
				ctx: context.Background(),
				model: &models.Model{
					ID: 1,
					Project: mlp.Project{
						ID: 1,
					},
					Type: models.ModelTypePyFuncV2,
				},
				version: &models.Version{
					ID: 1,
				},
				options: &models.BuildImageOptions{
					ResourceRequest: &models.ResourceRequest{
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
					},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewVersionImageService(tt.imageBuilder(), tt.imageBuilder())

			got, err := s.BuildImage(tt.args.ctx, tt.args.model, tt.args.version, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("versionImageService.BuildImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("versionImageService.BuildImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
