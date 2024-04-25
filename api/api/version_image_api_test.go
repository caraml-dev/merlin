package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/service/mocks"
)

func TestVersionImageController_GetImage(t *testing.T) {
	tests := []struct {
		name                string
		vars                map[string]string
		modelService        func() *mocks.ModelsService
		versionService      func() *mocks.VersionsService
		versionImageService func() *mocks.VersionImageService
		want                *Response
	}{
		{
			name: "success, image existed",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					ProjectID: models.ID(1),
					Type:      "pyfunc",
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
				}, nil)
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				svc.On("GetImage", mock.Anything, mock.Anything, mock.Anything).
					Return(models.VersionImage{
						ProjectID: 1,
						ModelID:   1,
						VersionID: 1,
						ImageRef:  "ghcr.io/caraml-dev/project-model:1",
						Exists:    true,
					}, nil)
				return svc
			},
			want: &Response{
				code: http.StatusOK,
				data: models.VersionImage{
					ProjectID: 1,
					ModelID:   1,
					VersionID: 1,
					ImageRef:  "ghcr.io/caraml-dev/project-model:1",
					Exists:    true,
				},
			},
		},
		{
			name: "failed, model not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("record not found"))
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				return svc
			},
			want: NotFound("Model not found: record not found"),
		},
		{
			name: "failed, invalid model type",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					ProjectID: models.ID(1),
					Type:      "xgboost",
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				return svc
			},
			want: BadRequest("Invalid model type: model type xgboost is not supported"),
		},
		{
			name: "failed, model version not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					ProjectID: models.ID(1),
					Type:      "pyfunc",
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(nil, fmt.Errorf("record not found"))
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				svc.On("GetImage", mock.Anything, mock.Anything, mock.Anything).
					Return(models.VersionImage{
						ProjectID: 1,
						ModelID:   1,
						VersionID: 1,
						ImageRef:  "ghcr.io/caraml-dev/project-model:1",
						Exists:    true,
					}, nil)
				return svc
			},
			want: NotFound("Model version not found: record not found"),
		},
		{
			name: "failed getting image",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					ProjectID: models.ID(1),
					Type:      "pyfunc",
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
				}, nil)
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				svc.On("GetImage", mock.Anything, mock.Anything, mock.Anything).
					Return(models.VersionImage{}, fmt.Errorf("timeout"))
				return svc
			},
			want: InternalServerError("Error getting image: timeout"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &VersionImageController{
				AppContext: &AppContext{
					ModelsService:       tt.modelService(),
					VersionsService:     tt.versionService(),
					VersionImageService: tt.versionImageService(),
				},
			}

			got := c.GetImage(&http.Request{}, tt.vars, nil)
			assertEqualResponses(t, tt.want, got)
		})
	}
}

func TestVersionImageController_BuildImage(t *testing.T) {
	tests := []struct {
		name                string
		vars                map[string]string
		body                interface{}
		modelService        func() *mocks.ModelsService
		versionService      func() *mocks.VersionsService
		versionImageService func() *mocks.VersionImageService
		want                *Response
	}{
		{
			name: "success, image existed",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			body: &models.BuildImageOptions{
				ResourceRequest: &models.ResourceRequest{
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					ProjectID: models.ID(1),
					Type:      "pyfunc",
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
				}, nil)
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				svc.On("BuildImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("ghcr.io/caraml-dev/project-model:1", nil)
				return svc
			},
			want: &Response{
				code: http.StatusAccepted,
			},
		},
		{
			name: "failed, model not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			body: &models.BuildImageOptions{
				ResourceRequest: &models.ResourceRequest{
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("record not found"))
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				return svc
			},
			want: NotFound("Model not found: record not found"),
		},
		{
			name: "failed, invalid model type",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			body: &models.BuildImageOptions{
				ResourceRequest: &models.ResourceRequest{
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					ProjectID: models.ID(1),
					Type:      "xgboost",
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				return svc
			},
			want: BadRequest("Invalid model type: model type xgboost is not supported"),
		},
		{
			name: "failed, model version not found",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			body: &models.BuildImageOptions{
				ResourceRequest: &models.ResourceRequest{
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					ProjectID: models.ID(1),
					Type:      "pyfunc",
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(nil, fmt.Errorf("record not found"))
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				return svc
			},
			want: NotFound("Model version not found: record not found"),
		},
		{
			name: "fail, bad request",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			body: nil,
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					ProjectID: models.ID(1),
					Type:      "pyfunc",
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
				}, nil)
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				svc.On("BuildImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("ghcr.io/caraml-dev/project-model:1", nil)
				return svc
			},
			want: BadRequest("Unable to parse request body"),
		},
		{
			name: "failed building image, but we still return Accepted",
			vars: map[string]string{
				"model_id":   "1",
				"version_id": "1",
			},
			body: &models.BuildImageOptions{
				ResourceRequest: &models.ResourceRequest{
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
				},
			},
			modelService: func() *mocks.ModelsService {
				svc := &mocks.ModelsService{}
				svc.On("FindByID", mock.Anything, models.ID(1)).Return(&models.Model{
					ID:        models.ID(1),
					ProjectID: models.ID(1),
					Type:      "pyfunc",
				}, nil)
				return svc
			},
			versionService: func() *mocks.VersionsService {
				svc := &mocks.VersionsService{}
				svc.On("FindByID", mock.Anything, models.ID(1), models.ID(1), mock.Anything).Return(&models.Version{
					ID:      models.ID(1),
					ModelID: models.ID(1),
				}, nil)
				return svc
			},
			versionImageService: func() *mocks.VersionImageService {
				svc := &mocks.VersionImageService{}
				svc.On("BuildImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("", fmt.Errorf("error building image: timeout"))
				return svc
			},
			want: &Response{
				code: http.StatusAccepted,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &VersionImageController{
				AppContext: &AppContext{
					ModelsService:       tt.modelService(),
					VersionsService:     tt.versionService(),
					VersionImageService: tt.versionImageService(),
				},
			}

			got := c.BuildImage(&http.Request{}, tt.vars, tt.body)
			assertEqualResponses(t, tt.want, got)
		})
	}
}
