/*
 * @Author: morehao morehao@qq.com
 * @Date: 2025-12-03 20:11:38
 * @LastEditors: morehao morehao@qq.com
 * @LastEditTime: 2025-12-04 11:18:10
 * @FilePath: /roc/apps/kecore/services/messagecenter/dto.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package messagecenter

import "github.com/insmtx/corekg/apps/kecore/models/foresttype"

type SendMessageReq struct {
	CompanyID     uint                           `json:"company_id"`
	UserID        uint                           `json:"user_id"`
	Uin           uint                           `json:"uin"`
	TemplateName  foresttype.MessageTemplateName `json:"template_name"`
	SourceType    foresttype.MessageSourceType   `json:"source_type"`
	SourceID      uint                           `json:"source_id"`
	MessageParams map[string]string              `json:"message_params"`
}

type SendMessageResp struct {
	MessageID uint `json:"message_id"`
}

type BatchSendMessageResp struct {
	MessageIDs []uint `json:"message_ids"`
}
