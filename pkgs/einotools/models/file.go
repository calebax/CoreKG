package models

import "encoding/json"

// File describes an input file available to an agent.
type File struct {
	// FileName is the display name of the input file.
	FileName string `json:"file_name"`
	// Type identifies the file category, such as image, PDF, document, CSV, audio, video, or URL.
	Type string `json:"type"`
	// FileSize is the source file size in bytes.
	FileSize int64 `json:"file_size"`
	// FileURL is the source or parsed-content URL available to tools.
	FileURL string `json:"file_url"`
	// FileOssUrl is retained for source compatibility with older callers.
	// Deprecated: use FileURL instead.
	FileOssUrl string `json:"-"`
	// Content contains file content that has already been read for direct model input.
	Content string `json:"content,omitempty"`
	// Truncated reports whether Content contains only a prefix of the source file.
	Truncated bool `json:"truncated,omitempty"`
	// Description explains how the agent should interpret the file and its content.
	Description string `json:"description"`
}

// MarshalJSON serializes FileURL and falls back to the deprecated FileOssUrl field for older callers.
func (f File) MarshalJSON() ([]byte, error) {
	type fileAlias File
	fileURL := f.FileURL
	if fileURL == "" {
		fileURL = f.FileOssUrl
	}
	return json.Marshal(struct {
		fileAlias
		FileURL string `json:"file_url"`
	}{
		fileAlias: fileAlias(f),
		FileURL:   fileURL,
	})
}
