package ragtask

import "github.com/insmtx/corekg/pkgs/task"

// AnalysisPDFTaskPayload is the payload for the AnalysisPDF task.
type AnalysisPDFTaskPayload struct {
	task.CommonPayload `json:",inline"`
	FileID             uint   `json:"file_id"`  // The ID of the file to extract text from
	FileURL            string `json:"file_url"` // The URL of the file to extract text from
	SubjectID          uint   `json:"subject_id"`
	UploadPath         string `json:"upload_path"` // file or dir
	CompanyID          uint   `json:"company_id"`
	ForestID           uint   `json:"forest_id"`
	Uin                uint   `json:"uin"`
	Bucket             string `json:"bucket"`
	ESIndex            string `json:"es_index"` // es 索引
}
