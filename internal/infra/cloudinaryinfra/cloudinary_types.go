package cloudinaryinfra

type UploadSignatureResponse struct {
	Signature    string `json:"signature"`
	Timestamp    string `json:"timestamp"`
	APIKey       string `json:"api_key"`
	CloudName    string `json:"cloud_name"`
	Folder       string `json:"folder"`
	PublicID     string `json:"public_id,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
}
