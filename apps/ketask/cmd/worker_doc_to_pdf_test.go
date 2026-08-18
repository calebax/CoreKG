package main

import (
	"context"
	"testing"

	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/stretchr/testify/assert"
)

func Test_doc2PDFWorker_Run(t *testing.T) {
	initResourceMap := map[string]struct{}{
		initResourceStorage: {},
	}
	err := resourceInit(initResourceMap)
	assert.Nil(t, err)

	testCases := []struct {
		name    string
		taskID  uint
		payload *ragtask.TaskPayload
	}{
		{
			name:   "docx",
			taskID: 31570,
			payload: &ragtask.TaskPayload{
				CommonPayload: task.CommonPayload{
					TaskType: "ke.doc_to_pdf_task",
					Timeout:  9000000000000,
				},
				CompanyID:     2,
				Uin:           384,
				ForestID:      873,
				FileID:        7835,
				Filename:      "2025110102.docx",
				FileURL:       "https://example.com:58081/test-knownow/forest/20251101/384-LGbbojfNB.docx",
				PreviewFileID: 37794,
				SubjectID:     7835,
				UploadPath:    "forest/20251101/384-5JF2SINfD.pdf",
				StoragePath:   "forest/20251101/384-LGbbojfNB.docx",
				Bucket:        "test-knownow",
				FileExt:       global.FileExtDOCX,
			},
		},
		// TODO 修改真实UploadPath和StoragePath
		// {
		// 	name:   "ofd",
		// 	taskID: 31571,
		// 	payload: &ragtask.TaskPayload{
		// 		CommonPayload: task.CommonPayload{
		// 			TaskType: "ke.doc_to_pdf_task",
		// 			Timeout:  9000000000000,
		// 		},
		// 		CompanyID:     2,
		// 		Uin:           384,
		// 		ForestID:      873,
		// 		FileID:        7835,
		// 		Filename:      "1273e542ee6d48c584d728d6c242d81f.ofd",
		// 		FileURL:       "https://example.com:58081/dotpen-test/apigateway/317/2026/3/16/1273e542ee6d48c584d728d6c242d81f/1273e542ee6d48c584d728d6c242d81f.ofd",
		// 		PreviewFileID: 37794,
		// 		SubjectID:     7835,
		// 		UploadPath:    "forest/20251101/384-1273e542ee6d48c584d728d6c242d81f.pdf",
		// 		StoragePath:   "forest/20251101/384-1273e542ee6d48c584d728d6c242d81f.ofd",
		// 		Bucket:        "test-knownow",
		// 		FileExt:       global.FileExtOFD,
		// 	},
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := newDoc2PDFWorker(context.Background(), tc.taskID, tc.payload).Run()
			assert.Nil(t, err)
		})
	}
}
