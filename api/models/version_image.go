package models

type BuildImageOptions struct {
	ResourceRequest *ResourceRequest `json:"resource_request,omitempty"`
}

type ImageBuildingJobStatus string

const (
	ImageBuildingJobStatusActive    ImageBuildingJobStatus = "active"
	ImageBuildingJobStatusSucceeded ImageBuildingJobStatus = "succeeded"
	ImageBuildingJobStatusFailed    ImageBuildingJobStatus = "failed"
	ImageBuildingJobStatusUnknown   ImageBuildingJobStatus = "unknown"
)

type VersionImage struct {
	ProjectID ID                     `json:"project_id"`
	ModelID   ID                     `json:"model_id"`
	VersionID ID                     `json:"version_id"`
	ImageRef  string                 `json:"image_ref"`
	Exists    bool                   `json:"exists"`
	JobStatus ImageBuildingJobStatus `json:"image_building_job_status"`
}
