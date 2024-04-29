package models

type BuildImageOptions struct {
	BackoffLimit    *int32           `json:"backoff_limit,omitempty"`
	ResourceRequest *ResourceRequest `json:"resource_request,omitempty"`
}

type ImageBuildingJobStatus struct {
	State   ImageBuildingJobState `json:"state"`
	Message string                `json:"message,omitempty"`
}

type ImageBuildingJobState string

const (
	ImageBuildingJobStateActive    ImageBuildingJobState = "active"
	ImageBuildingJobStateSucceeded ImageBuildingJobState = "succeeded"
	ImageBuildingJobStateFailed    ImageBuildingJobState = "failed"
	ImageBuildingJobStateUnknown   ImageBuildingJobState = "unknown"
)

type VersionImage struct {
	ProjectID ID                     `json:"project_id"`
	ModelID   ID                     `json:"model_id"`
	VersionID ID                     `json:"version_id"`
	ImageRef  string                 `json:"image_ref"`
	Exists    bool                   `json:"exists"`
	JobStatus ImageBuildingJobStatus `json:"image_building_job_status"`
}
