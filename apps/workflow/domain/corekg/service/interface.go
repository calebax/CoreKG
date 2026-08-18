package service

type CoreKGKnowledge interface {
	GetKnowledgeFiles(request *GetKnowledgeFilesRequest) (response *GetKnowledgeFilesResponse, err error)
	GetKnowledgeFilesUrl(request *GetKnowledgeFilesUrlRequest) (response *GetKnowledgeFilesUrlResponse, err error)
	GetKnowledgeSlice(request *GetKnowledgeSliceRequest) (response *GetKnowledgeSliceResponse, err error)
}

type GetKnowledgeFilesRequest struct {
	CoreKGToken       string
	CoreKGKnowledgeID uint
}

type GetKnowledgeFilesResponse struct {
	Response struct {
		Data []KnowledgeFile `json:"data"`
	} `json:"Response"`
}

type KnowledgeFile struct {
	Name   string `json:"name"`
	ID     uint   `json:"ID"`
	Ext    string `json:"ext"`
	Size   uint   `json:"size"`
	Status string `json:"status"`
}

type GetKnowledgeFilesUrlRequest struct {
	FileID      uint
	CoreKGToken string
}

type GetKnowledgeFilesUrlResponse struct {
	Response struct {
		Url string `json:"url"`
	} `json:"Response"`
}

type GetKnowledgeSliceRequest struct {
	FileID      uint
	ForestID    uint
	CoreKGToken string
}

type GetKnowledgeSliceResponse struct {
	Response struct {
		Chunks []struct {
			Source struct {
				Description string `json:"description"`
				Sequence    uint   `json:"sequence"`
			} `json:"_source"`
		} `json:"chunks"`
	} `json:"Response"`
}
