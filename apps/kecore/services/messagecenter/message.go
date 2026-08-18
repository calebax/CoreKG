/*
 * @Author: morehao morehao@qq.com
 * @Date: 2025-12-03 19:56:44
 * @LastEditors: morehao morehao@qq.com
 * @LastEditTime: 2025-12-04 11:19:28
 * @FilePath: /roc/apps/kecore/services/messagecenter/message.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package messagecenter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/sync/errgroup"
)

type Message struct {
	render *TemplateRender
}

func NewMessage() *Message {
	return &Message{
		render: NewTemplateRender(),
	}
}

func (m *Message) SendMessage(ctx context.Context, req *SendMessageReq) (*SendMessageResp, error) {
	if req.TemplateName == "" {
		return nil, fmt.Errorf("template name is empty")
	}
	if req.CompanyID == 0 {
		return nil, fmt.Errorf("company id is empty")
	}
	if req.UserID == 0 {
		return nil, fmt.Errorf("user id is empty")
	}
	if req.Uin == 0 {
		return nil, fmt.Errorf("uin is empty")
	}
	templateEntity, err := forest.NewKeMessageTemplateDao().GetByCond(ctx, &forest.KeMessageTemplateCond{
		Name: req.TemplateName,
	})
	if err != nil {
		return nil, fmt.Errorf("get message template fail, send req: %s, err: %v", logs.JSON(req), err)
	}
	if templateEntity == nil || templateEntity.ID == 0 {
		return nil, fmt.Errorf("message template not found, send req: %s, err: %v", logs.JSON(req), err)
	}

	title, err := m.render.Render(templateEntity.TitleTemplate, req.MessageParams)
	if err != nil {
		return nil, fmt.Errorf("render message title fail, send req: %s, err: %v", logs.JSON(req), err)
	}
	content, err := m.render.Render(templateEntity.ContentTemplate, req.MessageParams)
	if err != nil {
		return nil, fmt.Errorf("render message content fail, send req: %s, err: %v", logs.JSON(req), err)
	}
	routePath, err := m.render.Render(templateEntity.RoutePath, req.MessageParams)
	if err != nil {
		return nil, fmt.Errorf("render message route path fail, send req: %s, err: %v", logs.JSON(req), err)
	}
	messageEntity := &foresttype.KeUinMessage{
		CompanyID:    req.CompanyID,
		UserID:       req.UserID,
		Uin:          req.Uin,
		Title:        title,
		TemplateID:   templateEntity.ID,
		TemplateType: templateEntity.Type,
		Content:      content,
		SourceType:   req.SourceType,
		SourceID:     req.SourceID,
		RoutePath:    routePath,
		ReadStatus:   foresttype.MessageReadStatusUnread,
	}
	if err := forest.NewKeUinMessageDao().Insert(ctx, messageEntity); err != nil {
		return nil, fmt.Errorf("create user message fail, send req: %s, err: %v", logs.JSON(req), err)
	}
	return &SendMessageResp{MessageID: messageEntity.ID}, nil
}

func (m *Message) BatchSendMessage(ctx context.Context, reqList []*SendMessageReq) (*BatchSendMessageResp, error) {
	if len(reqList) == 0 {
		return &BatchSendMessageResp{MessageIDs: []uint{}}, nil
	}

	for i, v := range reqList {
		if v.TemplateName == "" {
			return nil, fmt.Errorf("template name is empty at index %d", i)
		}
		if v.CompanyID == 0 {
			return nil, fmt.Errorf("company id is empty at index %d", i)
		}
		if v.UserID == 0 {
			return nil, fmt.Errorf("user id is empty at index %d", i)
		}
		if v.Uin == 0 {
			return nil, fmt.Errorf("uin is empty at index %d", i)
		}
	}

	templateNameMap := make(map[foresttype.MessageTemplateName]struct{})
	var templateNameList []foresttype.MessageTemplateName
	for _, v := range reqList {
		templateNameMap[v.TemplateName] = struct{}{}
	}
	for k := range templateNameMap {
		templateNameList = append(templateNameList, k)
	}

	templateEntityList, err := forest.NewKeMessageTemplateDao().GetListByCond(ctx, &forest.KeMessageTemplateCond{
		NameList: templateNameList,
	})
	if err != nil {
		return nil, fmt.Errorf("get message template fail, template name list: %s, err: %v", logs.JSON(templateNameList), err)
	}
	if len(templateEntityList) != len(templateNameList) {
		return nil, fmt.Errorf("message template not found, template name list: %s", logs.JSON(templateNameList))
	}

	templateMap := templateEntityList.ToNameMap()

	messageEntityList := make(foresttype.KeUinMessageList, 0, len(reqList))
	for i, req := range reqList {
		templateEntity := templateMap[req.TemplateName]

		title, err := m.render.Render(templateEntity.TitleTemplate, req.MessageParams)
		if err != nil {
			return nil, fmt.Errorf("render message title fail at index %d, send req: %s, err: %v", i, logs.JSON(req), err)
		}

		content, err := m.render.Render(templateEntity.ContentTemplate, req.MessageParams)
		if err != nil {
			return nil, fmt.Errorf("render message content fail at index %d, send req: %s, err: %v", i, logs.JSON(req), err)
		}

		routePath, err := m.render.Render(templateEntity.RoutePath, req.MessageParams)
		if err != nil {
			return nil, fmt.Errorf("render message route path fail at index %d, send req: %s, err: %v", i, logs.JSON(req), err)
		}

		// 创建消息实体
		messageEntity := &foresttype.KeUinMessage{
			CompanyID:    req.CompanyID,
			UserID:       req.UserID,
			Uin:          req.Uin,
			Title:        title,
			Content:      content,
			TemplateID:   templateEntity.ID,
			TemplateType: templateEntity.Type,
			SourceType:   req.SourceType,
			SourceID:     req.SourceID,
			RoutePath:    routePath,
			ReadStatus:   foresttype.MessageReadStatusUnread,
		}
		messageEntityList = append(messageEntityList, *messageEntity)
	}

	// 按每 50 个一组进行分组
	const batchSize = 50
	batches := make([][]foresttype.KeUinMessage, 0, (len(messageEntityList)+batchSize-1)/batchSize)
	for i := 0; i < len(messageEntityList); i += batchSize {
		end := i + batchSize
		if end > len(messageEntityList) {
			end = len(messageEntityList)
		}
		batches = append(batches, messageEntityList[i:end])
	}

	// 使用 errgroup 并发创建
	g, gCtx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	messageIDs := make([]uint, 0, len(messageEntityList))

	for i := range batches {
		batch := batches[i] // 使用索引访问，避免闭包问题
		g.Go(func() error {
			// 检查上下文是否已取消
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			batchList := make(foresttype.KeUinMessageList, len(batch))
			copy(batchList, batch)

			if err := forest.NewKeUinMessageDao().BatchInsert(gCtx, batchList); err != nil {
				return fmt.Errorf("batch create user message fail, batch size: %d, err: %v", len(batchList), err)
			}

			// 收集插入后的 message IDs
			mu.Lock()
			for _, v := range batchList {
				messageIDs = append(messageIDs, v.ID)
			}
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("batch create user message fail, send req list: %s, err: %v", logs.JSON(reqList), err)
	}

	return &BatchSendMessageResp{MessageIDs: messageIDs}, nil
}

func (m *Message) MarkAsRead(ctx context.Context, messageID uint) error {
	if messageID == 0 {
		return fmt.Errorf("message id is empty")
	}
	messageEntity, err := forest.NewKeUinMessageDao().GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("get message by id fail, message id: %d, err: %v", messageID, err)
	}
	if messageEntity == nil || messageEntity.ID == 0 {
		return fmt.Errorf("message not found, message id: %d", messageID)
	}
	updateMap := map[string]interface{}{
		"read_status": foresttype.MessageReadStatusRead,
		"read_at":     time.Now(),
	}
	if err := forest.NewKeUinMessageDao().UpdateMap(ctx, messageID, updateMap); err != nil {
		return fmt.Errorf("mark message as read fail, message id: %d, err: %v", messageID, err)
	}
	return nil
}
