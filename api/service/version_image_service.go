package service

import (
	"context"
	"fmt"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/imagebuilder"
)

type VersionImageService interface {
	GetImage(ctx context.Context, model *models.Model, version *models.Version) (models.VersionImage, error)
	BuildImage(ctx context.Context, model *models.Model, version *models.Version, options *models.BuildImageOptions) (string, error)
}

func NewVersionImageService(webserviceBuilder imagebuilder.ImageBuilder, predJobBuilder imagebuilder.ImageBuilder) VersionImageService {
	return &versionImageService{
		webserviceBuilder: webserviceBuilder,
		predJobBuilder:    predJobBuilder,
	}
}

type versionImageService struct {
	webserviceBuilder imagebuilder.ImageBuilder
	predJobBuilder    imagebuilder.ImageBuilder
}

func (s *versionImageService) getImageBuilder(modelType string) (imagebuilder.ImageBuilder, error) {
	switch modelType {
	case models.ModelTypePyFunc:
		return s.webserviceBuilder, nil
	case models.ModelTypePyFuncV2:
		return s.predJobBuilder, nil
	default:
		return nil, fmt.Errorf("model type %s is not supported", modelType)
	}
}

func (s *versionImageService) GetImage(ctx context.Context, model *models.Model, version *models.Version) (models.VersionImage, error) {
	imageBuilder, err := s.getImageBuilder(model.Type)
	if err != nil {
		return models.VersionImage{}, err
	}

	versionImage := imageBuilder.GetVersionImage(ctx, model.Project, model, version)

	jobStatus := imageBuilder.GetImageBuildingJobStatus(ctx, model.Project, model, version)
	versionImage.JobStatus = jobStatus

	return versionImage, nil
}

func (s *versionImageService) BuildImage(ctx context.Context, model *models.Model, version *models.Version, options *models.BuildImageOptions) (string, error) {
	imageBuilder, err := s.getImageBuilder(model.Type)
	if err != nil {
		return "", err
	}

	imageRef, err := imageBuilder.BuildImage(ctx, model.Project, model, version, options.ResourceRequest, options.BackoffLimit)
	if err != nil {
		return "", fmt.Errorf("error building image: %w", err)
	}

	return imageRef, nil
}
