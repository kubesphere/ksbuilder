package cloud

import (
	"time"
)

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

type Snapshot struct {
	SnapshotID string `json:"snapshot_id"`
	Metadata   struct {
		Version string `json:"version"`
	} `json:"metadata"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Extension struct {
	ExtensionID   string `json:"extension_id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	LatestVersion struct {
		Version string `json:"version"`
	} `json:"latest_version"`
}

type ListExtensionsResponse struct {
	Extensions []Extension `json:"extensions"`
}
