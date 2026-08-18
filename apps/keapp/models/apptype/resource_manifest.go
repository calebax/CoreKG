package apptype

type ResourceManifest struct {
	ResourceID   string         `json:"resource_id"`
	ResourceType string         `json:"resource_type"`
	SourceURL    string         `json:"source_url"`
	ContentUnits []ContentUnit  `json:"content_units"`
	Metadata     map[string]any `json:"metadata"`
}

type ContentUnit struct {
	Type     string         `json:"type"`
	Location ContentLocator `json:"location"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

type ContentLocator struct {
	Selector string `json:"selector,omitempty"`
	Index    int    `json:"index,omitempty"`
	Title    string `json:"title,omitempty"`
}

type Evidence struct {
	ResourceID  string          `json:"resource_id"`
	ContentType string          `json:"content_type"`
	Score       float64         `json:"score"`
	Locator     EvidenceLocator `json:"locator"`
	Snippet     string          `json:"snippet"`
	Payload     map[string]any  `json:"payload"`
}

type EvidenceLocator struct {
	URL      string `json:"url,omitempty"`
	Title    string `json:"title,omitempty"`
	Selector string `json:"selector,omitempty"`
}

type IndexArtifact struct {
	IndexType  string `json:"index_type"`
	DocumentID string `json:"document_id"`
	Payload    any    `json:"payload"`
}

type IndexBuilder interface {
	Build(manifest ResourceManifest) ([]IndexArtifact, error)
	IndexType() string
}
