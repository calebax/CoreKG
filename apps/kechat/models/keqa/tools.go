package keqa

import (
	"context"
	"encoding/json"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
)

// TransformChatReferenceList 转换聊天参考列表
func TransformChatReferenceList(cl chattype.QueryReferenceList) ([]byte, error) {
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
			Filename:  cr.FileName,
			ChunkList: make([]Chunk, 0, len(cr.ChunkList)),
		}
		for _, chunk := range cr.ChunkList {
			if version.DeployMode() != "" || version.DeployMode() != global.DeployModeOpenPO {
				chunk.ImageURL = fs.SplitHost(context.TODO(), chunk.ImageURL)
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

func RewriteChatReferenceImageURLs(
	ctx context.Context,
	refs chattype.QueryReferenceList,
) (chattype.QueryReferenceList, error) {
	mode := version.DeployMode()
	if mode == "" || mode == global.DeployModeOpenPO {
		return refs, nil
	}

	for i := range refs {
		for j := range refs[i].ChunkList {
			u := refs[i].ChunkList[j].ImageURL
			if u == "" {
				continue
			}
			refs[i].ChunkList[j].ImageURL = fs.SplitHost(context.TODO(), u)
		}
	}
	out := refs

	return out, nil
}
