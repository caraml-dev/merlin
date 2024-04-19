package models

type BuildImageOptions struct {
	ResourceRequest *ResourceRequest `json:"resource_request,omitempty"`
}

type VersionImage struct {
	ProjectID ID     `json:"project_id"`
	ModelID   ID     `json:"model_id"`
	VersionID ID     `json:"version_id"`
	ImageRef  string `json:"image_ref"`
	Existed   bool   `json:"existed"`
}
