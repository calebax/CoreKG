package apis

import (
	"github.com/insmtx/corekg/apps/account/accountmds"
	"github.com/insmtx/corekg/apps/kechat/mds"
	"github.com/insmtx/corekg/apps/kecore/services/deployhandle"
	"github.com/insmtx/corekg/apps/kellm"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/metrics"
)

// RegistryRouter 注册路由
func RegistryRouter(eng *server.Router) {
	eng.HandleDoc("kechat")
	eng.G("chat.metrics", metrics.GetHandler())
	eng.AuthInject("yygu", accountmds.InjectLoginStatus)
	eng.AuthInject("", accountmds.InjectLoginStatus)

	eng.P("chat.Agent/chat/completions", accountmds.RequireAgentChatPrivilege, mds.QAQuotaMD, AgentChat)
	// 聊天管理
	eng.PRequireLogin("chat.SubmitChatQuestionStream", mds.QAQuotaMD, SubmitChatQuestionStream)
	eng.PRequireLogin("chat.ChatQuestionStream", mds.QAQuotaMD, ChatQuestionStream)
	eng.PRequireLogin("chat.NewChatQuestionStream", NewChatQuestionStream)
	eng.PRequireLogin("chat.NewChatSession", NewChatSession)
	eng.PRequireLogin("chat.RenameChatSession", RenameChatSession)
	eng.PRequireLogin("chat.RemoveChatSession", RemoveChatSession)
	eng.PRequireLogin("chat.ListChatSession", ListChatSession)
	eng.PRequireLogin("chat.ListSessionChats", ListSessionChats)
	eng.PRequireLogin("chat.SetTopChatSession", SetTopChatSession)
	eng.PRequireLogin("chat.GetSessionInfo", GetSessionInfo)
	eng.PRequireLogin("chat.GetQuestionInfo", GetQuestionInfo)
	eng.PRequireLogin("chat.WorkflowTestRun", WorkflowTestRun)
	eng.PRequireLogin("chat.UpdateChatAgentInput", UpdateChatAgentInput)
	// 更新工作流版本
	eng.PRequireLogin("chat.UpdateWorkflowVersion", UpdateWorkflowVersion)

	eng.PRequireLogin("chat.GetFileSession", GetFileSession)

	eng.PRequireLogin("chat.ChatGetMessage", ChatGetMessage)
	eng.PRequireLogin("chat.StopChat", StopChat)
	eng.PRequireLogin("chat.SubmitChatQuestion", SubmitChatQuestion)

	// Agent
	eng.PRequireLogin("chat.ListChatAgent", ListChatAgent)
	eng.PRequireLogin("chat.UpdateChatAgent", UpdateChatAgent)
	eng.PRequireLogin("chat.DeleteChatAgent", DeleteChatAgent)
	eng.PRequireLogin("chat.GetAgentInfo", mds.HasAgentUsePerm, GetAgentInfo)
	eng.PRequireLogin("chat.CreateAgentApp", CreateAgentApp)
	eng.PRequireLogin("chat.ListCollectApp", ListCollectApp)
	eng.PRequireLogin("chat.CollectApp", CollectApp)

	// agent管理
	eng.PRequireLogin("chat.CreateAgent", CreateAgent)
	eng.PRequireLogin("chat.UpdateAgent", UpdateAgent)
	eng.PRequireLogin("chat.GetAgentDetail", GetAgentDetail)
	eng.PRequireLogin("chat.TestAgent", TestAgent)
	// 机器人版本管理
	eng.PRequireLogin("chat.ListAgentVersion", ListAgentVersion)
	eng.PRequireLogin("chat.ChooseAgentVersion", ChooseAgentVersion)

	// 模型管理
	eng.PRequireLogin("chat.ListModel", ListModel)
	eng.PRequireLogin("chat.CreateModel", accountmds.DecryptMD("request.api_key"), CreateModel)
	eng.PRequireLogin("chat.DeleteModel", DeleteModel)
	eng.PRequireLogin("chat.UpdateModel", accountmds.DecryptMD("request.api_key"), UpdateModel)
	eng.PRequireLogin("chat.GetModelDetail", GetModelDetail)
	eng.PRequireLogin("chat.ModelTest", accountmds.DecryptMD("request.api_key"), ModelTest)
	eng.PRequireLogin("chat.BindCozeModel", BindCozeModel)

	// 上传图片
	eng.PRequireLogin("chat.UploadImage", UploadImage)

	// 上传chat解析附件
	eng.PRequireLogin("chat.UploadAttachment", UploadAttachment)

	// Agent 权限
	eng.PRequireLogin("chat.GetAgentPermSet", GetAgentPermSet)
	eng.PRequireLogin("chat.ModifyChatPermSet", ModifyChatPermSet)

	// External api
	eng.P("chat.ExternalToken", ExternalToken)
	eng.PRequireLogin("chat.GetPersonalAccessToken", GetPersonalAccessToken)
	eng.PRequireLogin("chat.SetExternalStatus", SetExternalStatus)
	eng.PRequireLogin("chat.GetExternalStatus", GetExternalStatus)

	eng.P("chat.NewExternalChatStream", mds.CozeAgentAuthMD, mds.CozeConversationAuthMD, NewExternalChatStream)
	eng.P("chat.SubmitExternalChatStream", mds.CozeAgentAuthMD, mds.CozeConversationAuthMD, SubmitExternalChatStream)
	eng.P("chat.ChatAgent/chat/completions", mds.CozeAgentAuthMD, CozeAgentChat)

	//新权限
	eng.PRequireLogin("chat.GetAgentWithPerm", mds.HasAgentManagePerm, GetAgentWithPerm)
	eng.PRequireLogin("chat.UpdateAgentWithPerm", mds.AgentQuotaMD, UpdateAgentWithPerm)

	// 单文档智慧问答
	eng.PRequireLogin("chat.ListFileQA", ListFileQA)
	eng.PRequireLogin("chat.FileChat", mds.QAQuotaMD, FileChat)
	eng.PRequireLogin("chat.DeleteFileQA", DeleteFileQA)

	eng.P("chat.SaveChartCanvas", SaveChartCanvas)
	eng.P("chat.GetChartCanvas", GetChartCanvas)
	eng.P("chat.BatchDeleteChart", BatchDeleteChart)

	// 数据迁移
	eng.P("chat.MigrateChatQuestion", MigrateChatQuestion)
	eng.P("chat.MigrateForestChat", MigrateForestChat)

	// 创建coze插件
	eng.PRequireLogin("chat.CreateCozePlugin", CreateCozePlugin)
	// 删除coze映射
	eng.P("chat.DeleteCozeMappingByCozeID", DeleteCozeMappingByCozeID)
	// 获取智能体映射关系
	eng.PRequireLogin("chat.GetAgentMapping", GetAgentMapping)

	// 获取最近使用智能体
	eng.PRequireLogin("chat.GetLatestAgents", GetLatestAgents)
	//	移动Session
	eng.PRequireLogin("chat.MoveSession", MoveSession)

	eng.P("chat.SwitchPrivateEvn", deployhandle.SwitchPrivateEvn)
	eng.P("chat.NowDeployMode", deployhandle.NowDeployMode)

	eng.PRequireLogin("chat.GetAgentQuestionExcel", GetAgentQuestionExcel)
	eng.PRequireLogin("chat.GetAgentQuestionCount", GetAgentQuestionCount)

	// 问题扩写
	eng.PRequireLogin("chat.ExpansionQuestion", ExpansionQuestion)

	// LLM 透传
	eng.P("chat.LLM/chat/completions", kellm.ChatCompletions)
}
