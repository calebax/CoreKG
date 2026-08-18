package keqa

import (
	"context"
	"encoding/json"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/version"
)

// TransformChatReferenceList 转换聊天参考列表
func TransformChatReferenceList(ctx context.Context, cl foresttype.ChatReferenceList) ([]byte, error) {
	// Chunk 引用内容
	type Chunk struct {
		Sequence int    `json:"sequence"`
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
	}
	// Reference 聊天的引用内容
	type Reference struct {
		Filename  string  `json:"filename"`
		ChunkList []Chunk `json:"chunk_list"`
	}
	retList := make([]Reference, 0, len(cl))
	for _, cr := range cl {
		newCr := Reference{
			Filename:  cr.Filename,
			ChunkList: make([]Chunk, 0, len(cr.ChunkList)),
		}
		for _, chunk := range cr.ChunkList {
			if version.DeployMode() != "" {
				chunk.ImageURL = fs.SplitHost(ctx, chunk.ImageURL)
			}
			newCr.ChunkList = append(newCr.ChunkList, Chunk{
				Sequence: chunk.Sequence,
				Content:  chunk.Content,
				ImageURL: chunk.ImageURL,
			})
		}
		retList = append(retList, newCr)
	}

	return json.Marshal(retList)
}
