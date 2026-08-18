package keqa

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

// CreateForestQA 创建一个知识库问题
func CreateForestQA(session *foresttype.KnownowQASession, question string, urlList []string) (*foresttype.KnownowForestQA, error) {
	qa := &foresttype.KnownowForestQA{
		SessionID:    session.ID,
		CompanyID:    session.CompanyID,
		Uin:          session.Uin,
		Question:     question,
		ImageUrlList: types.NewStringArray(urlList),
		Answer:       "",
		MindGraph:    "",
		ImageContent: "",
		Status:       foresttype.QAStatusPending,
	}
	if err := dbutil.Knownow().Create(qa).Error; err != nil {
		return nil, err
	}
	return qa, nil
}

// ForestChat 知识问答
func ForestChat(ctx *gin.Context, qs *foresttype.KnownowForestQA, session *foresttype.KnownowQASession) (*foresttype.KnownowForestQA, error) {
	logs.InfoContextf(ctx, "[ForestChat] ForestChat: %v,session: %v", qs.Question, session)
	if session.Name == foresttype.DefaultSessionName {
		runes := []rune(qs.Question)
		session.Name = qs.Question
		if len(runes) > 10 {
			session.Name = string(runes[:10])
		}
		logs.InfoContextf(ctx, "[ForestChat] session.Name: %v", session.Name)
		err := ModifySession(ctx, session)
		if err != nil {
			logs.ErrorContextf(ctx, "[ForestChat] Failed to ModifySession: %v", err)
		}
	}

	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	defer sseClient.Close(ctx, fmt.Sprintf("%v", qs.ID))
	sseClient.SetHeaders(ctx.Writer)

	history, err := GetForestChatHistory(ctx, qs)
	if err != nil {
		return nil, err
	}
	imagelist := qs.ImageUrlList.Slice()

	if len(imagelist) > 0 {
		// 解析多模态
		qs.ImageContent = fmt.Sprintf("\n用户上传了%v张图片，根据描述信息找出与问题相关的可用信息并用于对话中，每条图片的具体描述信息如下：\n", len(imagelist))
		for i, url := range imagelist {
			res, err := DoImageParseRequest(ctx, url)
			if err != nil {
				logs.ErrorContextf(ctx, "[ForestChat] Failed to parse image: %v", err)
			}
			qs.ImageContent += fmt.Sprintf("第%v张\n%v\n", i+1, res)
		}
	}

	wrapper, err := HandelSearchReference(ctx, session.ForestIDList.Slice(), session.FileIDList.Slice(), session.EsIndex, qs.Question+qs.ImageContent)
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to HandelSearchReference: %v", err)
		return nil, err
	}

	// 查找问答对
	fqa, err := wrapper.searchWrapper.FindFQAByQuestion()
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to FindFQAByQuestion: %v", err)
		return nil, err
	}
	if len(fqa.Hits.Hits) != 0 {
		logs.InfoContextf(ctx, "[ForestChat] FindFQAByQuestion result: %v", len(fqa.Hits.Hits))
		// 写入
		if stoped, err := sseClient.WriteMessage(ctx, ctx.Writer, fmt.Sprintf("%v", qs.ID), WriteResult{
			Content: fqa.Hits.Hits[0].Source.QAAnswer,
		}.String()); err != nil {
			logs.ErrorContextf(ctx, "[forestqa] Failed to write Thinking response to KEQA: %v", err)
		} else if stoped {
			logs.ErrorContextf(ctx, "[forestqa] stream Stoped by KEQA")
			return nil, fmt.Errorf("stream KEQA stoped")
		}
		qs.Answer = fqa.Hits.Hits[0].Source.QAAnswer
		qs.Status = foresttype.QAStatusAnswered
		if err := dbutil.Knownow().Save(qs).Error; err != nil {
			logs.ErrorContextf(ctx, "[ForestChat] Failed to save QA: %v", err)
			return nil, err
		}
		return qs, nil
	}

	// 开始意图识别，正常搜索还是总结性问题
	intention, err := wrapper.searchWrapper.IntentionRecognition()
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to IntentionRecognition: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "[ForestChat] IntentionRecognition result: %v", intention)
	var answer, reason string
	switch intention {
	case "C", "c":
		answer, reason, err = DescriptionChat(wrapper, sseClient, qs, history, session)
	default:
		answer, reason, err = DefaultChat(wrapper, sseClient, qs, history, session)
	}

	qs.Answer = answer
	qs.Reasoning = reason
	qs.Status = foresttype.QAStatusAnswered
	if err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to search answer: %v", err)
		qs.Status = foresttype.QAStatusFailed
	}
	if err := dbutil.Knownow().Save(qs).Error; err != nil {
		logs.ErrorContextf(ctx, "[ForestChat] Failed to save QA: %v", err)
		return nil, err
	}
	return qs, nil
}

// QueryForestQAListResponse 表单类型响应
type QueryForestQAListResponse struct {
	apiobj.QueryResponse

	Data []*foresttype.KnownowForestQA
}

type QueryForestQAListOption struct {
	Uin       uint `json:"uin"`
	SessionID uint `json:"session_id"`
}

// QueryForestQAList 查询表单类型列表
func QueryForestQAList(opt *QueryForestQAListOption, ret *QueryForestQAListResponse) error {
	ret.Data = []*foresttype.KnownowForestQA{}

	query := dbutil.Knownow().Table(foresttype.TableNameKnownowForestQA).
		Where("deleted_at is null").
		Where("uin = ?", opt.Uin).
		Where("session_id = ?", opt.SessionID)

	if err := query.Count(&ret.Total).Error; err != nil {
		return err
	}
	if ret.Total == 0 {
		return nil
	}

	err := query.Find(&ret.Data).Error
	if err != nil {
		return err
	}
	return nil
}

type Chat struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// GetForestChatHistory 获取chat历史记录
func GetForestChatHistory(ctx context.Context, qs *foresttype.KnownowForestQA) (string, error) {
	opt := &QueryForestQAListOption{
		Uin:       qs.Uin,
		SessionID: qs.SessionID,
	}
	out := &QueryForestQAListResponse{}
	err := QueryForestQAList(opt, out)
	if err != nil {
		logs.ErrorContextf(ctx, "QueryForestQAList failed ,err = %v", err)
		return "", err
	}
	chats := []*Chat{}

	for _, qa := range out.Data {
		if qa.Status != foresttype.QAStatusAnswered {
			continue
		}
		chats = append(chats, &Chat{
			Question: qa.Question,
			Answer:   qa.Answer,
		})
	}

	chatsJSON, err := json.Marshal(chats)
	if err != nil {
		return "", err
	}
	return string(chatsJSON), nil
}
