package cloud

type userInfo struct {
	UserID string `json:"user_id"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type fileUsageResponse struct {
	Dir string `json:"dir"`
}

type UploadFilesResponse struct {
	Files []struct {
		URL string `json:"url"`
	} `json:"files"`
}

type UploadExtensionResponse struct {
	Snapshot struct {
		SnapshotID string `json:"snapshot_id"`
	} `json:"snapshot"`
}
