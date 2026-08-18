package apis

import (
	"github.com/insmtx/corekg/apps/account/accountmds"
	chatmds "github.com/insmtx/corekg/apps/kechat/mds"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/accountctl"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/articlectl"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/fileparse"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/fileqactl"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/forestctl"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/forestqactl"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/golbalsearchctl"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/graphctl"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/nbqueue"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/projectctl"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/qapair"
	"github.com/insmtx/corekg/apps/kecore/mds"
	"github.com/insmtx/corekg/apps/kecore/services/deployhandle"
	paynotify "github.com/insmtx/corekg/apps/kesale/notify"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/metrics"
)

func RegistryRouter(eng *server.Router) {
	eng.HandleDoc("kecore")
	eng.G("forest.metrics", metrics.GetHandler())
	eng.AuthInject("dotpen", accountmds.InjectLoginStatus)
	eng.AuthInject("yygu", accountmds.InjectLoginStatus)

	//Create
	eng.PRequireLogin("forest.CreateForest", forestctl.CreateForest)
	//List
	eng.PRequireLogin("forest.ListForest", forestctl.ListForest)
	//Delete
	eng.PRequireLogin("forest.DeleteFile", forestctl.DeleteFile)
	eng.PRequireLogin("forest.DeleteForest", forestctl.DeleteForest)
	eng.PRequireLogin("forest.ModifyForest", forestctl.ModifyForest)
	eng.PRequireLogin("forest.GetForest", mds.HasForestUsePerm, forestctl.GetForest)
	eng.PRequireLogin("forest.ListFile", forestctl.ListFile)
	eng.PRequireLogin("forest.CreateDir", forestctl.CreateDir)
	eng.PRequireLogin("forest.DeletePath", forestctl.DeleteDir)

	// Deprecated
	eng.PRequireLogin("forest.UploadFile", mds.DiskQuotaMD, forestctl.UploadFile)

	eng.PRequireLogin("forest.RenamePath", forestctl.RenamePath)
	eng.PRequireLogin("forest.MovePath", forestctl.MovePath)
	eng.PRequireLogin("forest.PreviewFile", forestctl.PreviewFile)
	eng.PRequireLogin("forest.GetFileInfo", forestctl.GetFileInfo)
	eng.PRequireLogin("forest.PreviewFileByURL", forestctl.PreviewFileByURL)
	eng.PRequireLogin("forest.RecentlyForest", forestctl.RecentlyForest)
	eng.PRequireLogin("forest.GetFilePath", forestctl.GetFilePath)
	// 上传文件相关
	eng.PRequireLogin("forest.PreUploadFile", mds.DiskPreQuotaMDV2, PreUploadFile)
	eng.PRequireLogin("forest.UploadFileCallBack", UploadFileCallBack)
	eng.PRequireLogin("forest.AbortUpload", AbortUpload)
	eng.PRequireLogin("forest.RenewUploadUrl", RenewUploadUrl)

	eng.PRequireLogin("forest.GetNameByModuleIDs", forestctl.GetNameByModuleIDs)
	eng.PRequireLogin("forest.ListParseHistory", forestctl.ListParseHistory)
	eng.PRequireLogin("forest.RetryParse", forestctl.RetryParse)
	eng.PRequireLogin("forest.ResplitChunk", forestctl.ResplitChunk)
	eng.PRequireLogin("forest.GetResourceBaseInfo", forestctl.GetResourceBaseInfo)
	//rename forest
	eng.PRequireLogin("forest.RenameForest", forestctl.RenameForest)
	eng.PRequireLogin("forest.UpdateForestDescription", UpdateForestDescription)

	// 单文档智慧问答
	eng.PRequireLogin("forest.ListFileQA", fileqactl.ListFileQA)
	//单文档问答新增
	eng.PRequireLogin("forest.FileChat", chatmds.QAQuotaMD, fileqactl.FileChat)
	eng.PRequireLogin("forest.DeleteFileQA", fileqactl.DeleteFileQA)

	// 知识森林问答
	eng.PRequireLogin("forest.ListSession", forestqactl.ListSession)
	eng.PRequireLogin("forest.ListSessionQA", forestqactl.ListSessionQA)
	//新增问答数据
	eng.PRequireLogin("forest.CreateForestSession", mds.QAChatQuotaWithErrMD, forestqactl.CreateForestSession)
	eng.PRequireLogin("forest.CreateForestQA", mds.QAChatQuotaMD, forestqactl.CreateForestQA)
	eng.PRequireLogin("forest.ForestQAChat", mds.QAChatQuotaMD, forestqactl.ForestQAChat)
	eng.PRequireLogin("forest.ModifySession", forestqactl.ModifySession)
	eng.PRequireLogin("forest.DeleteSession", forestqactl.DeleteSession)
	eng.PRequireLogin("forest.StopChat", forestqactl.StopChat)
	eng.PRequireLogin("forest.ChatGetMessage", forestqactl.ChatGetMessage)
	eng.PRequireLogin("forest.GetSessionInfo", forestqactl.GetSessionInfo)

	// 图数据库相关
	eng.PRequireLogin("forest.GetForestWordCloud", nbqueue.GetForestWordCloud)
	eng.PRequireLogin("forest.GetNodesByID", nbqueue.GetNodesByID)
	eng.PRequireLogin("forest.GetForestWordCloudGraph", nbqueue.GetForestWordCloudGraph)

	// 知识库检索
	eng.PRequireLogin("forest.ForestSearch", golbalsearchctl.ForestSearch)
	eng.PRequireLogin("forest.ForestSearchDoc", golbalsearchctl.ForestSearchDoc)
	eng.PRequireLogin("forest.ForestSearchImage", golbalsearchctl.ForestSearchImage)
	eng.PRequireLogin("forest.ForestSearchVideo", golbalsearchctl.ForestSearchVideo)

	//async parse
	eng.PRequireLogin("forest.GetContent", fileparse.GetContent)
	eng.PRequireLogin("forest.GetAnalysis", fileparse.GetAnalysis)
	eng.PRequireLogin("forest.GetMindMap", fileparse.GetMindMap)
	eng.PRequireLogin("forest.GetFileDescription", fileparse.GetFileDescription)
	eng.PRequireLogin("forest.GetRecommendQuestions", fileparse.GetRecommendQuestions)

	//用户角色权限

	//获取用户权限集
	eng.PRequireLogin("forest.GetForestPermSet", accountctl.GetForestPermSet)
	eng.PRequireLogin("forest.ModifyForestPermSet", accountctl.ModifyForestPermSet)

	//知识库权限范围接口
	eng.PRequireLogin("forest.ListForestPublicScope", forestctl.ListForestPublicScope)
	eng.PRequireLogin("forest.UpdateForestPublicScope", forestctl.UpdateForestPublicScope)
	eng.PRequireLogin("forest.UpdateForestManager", forestctl.UpdateForestManager)

	// 全局搜索
	eng.PRequireLogin("forest.GlobalSearch", golbalsearchctl.GlobalSearch)
	eng.PRequireLogin("forest.GlobalSearchDoc", golbalsearchctl.GlobalSearchDoc)
	eng.PRequireLogin("forest.GlobalSearchImage", golbalsearchctl.GlobalSearchImage)
	eng.PRequireLogin("forest.GlobalSearchVideo", golbalsearchctl.GlobalSearchVideo)
	eng.PRequireLogin("forest.GlobalSearchAgent", golbalsearchctl.GlobalSearchAgent)
	eng.PRequireLogin("forest.GlobalSearchForest", golbalsearchctl.GlobalSearchForest)

	//qa pair
	eng.PRequireLogin("forest.CreateQAPair", qapair.CreateQAPair)
	eng.PRequireLogin("forest.DeleteQAPair", qapair.DeleteQAPair)
	eng.PRequireLogin("forest.ModifyQAPair", qapair.ModifyQAPair)
	eng.PRequireLogin("forest.ListQAPair", qapair.ListQAPair)
	//批量导入
	eng.PRequireLogin("forest.UploadQAPair", qapair.UploadQAPair)
	eng.PRequireLogin("forest.CommitQAPair", qapair.CommitQAPair)

	//团队资源配额
	eng.PRequireLogin("forest.GetCompanyQuota", accountctl.GetCompanyQuota)
	eng.P("forest.VersionUpgradeSendCode", accountctl.VersionUpgradeSendCode)
	eng.P("forest.VersionUpgradeVerify", accountctl.VersionUpgradeVerify)

	//新权限
	eng.PRequireLogin("forest.UpdateForestWithPerm", accountctl.UpdateForestWithPerm)

	// 表格问答
	eng.PRequireLogin("forest.ListExcelSheet", forestctl.ListExcelSheet)

	// 数据库知识库
	eng.PRequireLogin("forest.CreateForestDBInstance", accountmds.DecryptMD("request.username", "request.password"), forestctl.CreateForestDBInstance)
	eng.PRequireLogin("forest.TestForestDBInstanceConnection", accountmds.DecryptMD("request.username", "request.password"), forestctl.TestForestDBInstanceConnection)
	eng.PRequireLogin("forest.ModifyForestDBInstance", accountmds.DecryptMD("request.username", "request.password"), forestctl.ModifyForestDBInstance)
	eng.PRequireLogin("forest.GetForestDBInstance", forestctl.GetForestDBInstance)
	eng.PRequireLogin("forest.ListForestDB", forestctl.ListForestDB)
	eng.PRequireLogin("forest.ListForestTable", forestctl.ListForestTable)
	eng.PRequireLogin("forest.GetForestTableMetadata", forestctl.GetForestTableMetadata)

	// 新版知识图谱
	eng.PRequireLogin("forest.CreateGraph", mds.GraphQuotaMD, graphctl.CreateGraph)
	eng.PRequireLogin("forest.UpdateGraph", graphctl.UpdateGraph)
	eng.PRequireLogin("forest.DeleteGraph", graphctl.DeleteGraph)
	eng.PRequireLogin("forest.CreateTag", graphctl.CreateTag)
	eng.PRequireLogin("forest.CreateEdge", graphctl.CreateEdge)
	eng.PRequireLogin("forest.ListForestGraph", graphctl.ListForestGraph)
	eng.PRequireLogin("forest.GetGraphInfo", mds.HasGraphUsePerm, graphctl.GetGraphInfo)
	eng.PRequireLogin("forest.ListGraphTag", graphctl.ListGraphTag)
	eng.PRequireLogin("forest.ListGraphNode", graphctl.ListGraphNode)
	eng.PRequireLogin("forest.UpdateTag", graphctl.UpdateTag)
	eng.PRequireLogin("forest.GetKnowledgeGraph", graphctl.GetKnowledgeGraph)
	eng.PRequireLogin("forest.GetTagEdge", graphctl.GetTagEdge)
	eng.PRequireLogin("forest.SubmitTemplate", graphctl.SubmitTemplate)
	eng.PRequireLogin("forest.DeleteEdge", graphctl.DeleteEdge)
	eng.PRequireLogin("forest.DeleteTag", graphctl.DeleteTag)
	eng.PRequireLogin("forest.UpdateEdge", graphctl.UpdateEdge)
	eng.PRequireLogin("forest.CreateForestGraph", CreateForestGraph)

	eng.PRequireLogin("forest.ParseGraph", graphctl.ParseGraph)
	eng.PRequireLogin("forest.RestockGraph", graphctl.RestockGraph)
	eng.P("forest.GraphTaskCallback", graphctl.GraphTaskCallback)

	eng.PRequireLogin("forest.CreateProject", projectctl.CreateProject)
	eng.PRequireLogin("forest.GetProjectInfo", projectctl.GetProjectInfo)
	eng.PRequireLogin("forest.DeleteProject", projectctl.DeleteProject)
	eng.PRequireLogin("forest.RenameProject", projectctl.RenameProject)
	eng.PRequireLogin("forest.ListProject", projectctl.ListProject)
	eng.PRequireLogin("forest.GetDefaultProject", GetDefaultProject)

	// 写作空间
	// 文章增删改查
	eng.PRequireLogin("forest.ListArticle", ListArticle)
	eng.PRequireLogin("forest.CreateArticle", mds.ArticleQuotaMD, CreateArticle)
	eng.PRequireLogin("forest.EditArticle", EditArticle)
	eng.PRequireLogin("forest.DeleteArticle", DeleteArticle)
	//获取文档内容
	eng.PRequireLogin("forest.GetArticle", GetArticle)
	//保存文档内容
	eng.PRequireLogin("forest.SaveArticleContent", SaveArticleContent)
	//查询模板列表
	// Deprecated: 文章模板接口已合并，请使用对应 Article 接口 + type 参数
	eng.PRequireLogin("forest.ListTemplate", ListArticleTemplate)
	//获取模板内容
	// Deprecated: 文章模板接口已合并，请使用对应 Article 接口 + type 参数
	eng.PRequireLogin("forest.GetArticleTemplate", GetArticleTemplate)
	//复制文章
	eng.PRequireLogin("forest.DuplicateArticle", DuplicateArticle)
	//ai写作指令执行接口
	eng.PRequireLogin("forest.ExecuteAIWriteCmd", ExecuteAIWriteCmd)
	//ai撰写接口
	eng.PRequireLogin("forest.AIWrite", articlectl.AIWrite)
	//保存为写作模板
	eng.PRequireLogin("forest.SaveAsArticleTemplate", SaveAsArticleTemplate)
	//删除写作模板
	// Deprecated: 文章模板接口已合并，请使用对应 Article 接口 + type 参数
	eng.PRequireLogin("forest.DeleteArticleTemplate", DeleteArticleTemplate)
	//修改写作模板
	// Deprecated: 文章模板接口已合并，请使用对应 Article 接口 + type 参数
	eng.PRequireLogin("forest.ModifyArticleTemplate", ModifyArticleTemplate)
	//创建写作模板
	// Deprecated: 文章模板接口已合并，请使用对应 Article 接口 + type 参数
	eng.PRequireLogin("forest.CreateArticleTemplate", CreateArticleTemplate)

	eng.PRequireLogin("forest.SetResourceEnable", forestctl.SetResourceEnable)

	//获取project item列表
	eng.PRequireLogin("forest.ListProjectItem", ListProjectItem)
	//获取project item详情
	eng.PRequireLogin("forest.GetProjectItem", GetProjectItem)

	// * back door
	// 获取资源url列表
	eng.P("forest.GetOriginResource", mds.HasHeaderStr(mds.AuthHeaderKey, mds.AuthValueGetOriginResource), GetOriginResource)

	eng.P("forest.SwitchPrivateEvn", deployhandle.SwitchPrivateEvn)
	eng.P("forest.NowDeployMode", deployhandle.NowDeployMode)
	eng.PRequireLogin("forest.StatAlgoMarkdown", StatAlgoMarkdown)
	eng.PRequireLogin("forest.RewriteMarkdownURL", RewriteMarkdownURL)

	eng.PRequireLogin("forest.ListPackage", ListPackage)
	eng.PRequireLogin("forest.CreateOrder", CreateOrder)
	eng.PRequireLogin("forest.QueryOrderStatus", QueryOrderStatus)

	eng.PRequireLogin("forest.GetCommonInfo", GetCommonInfo)

	eng.PRequireLogin("forest.ListPaymentOrderRecord", ListPaymentOrderRecord)

	eng.P("forest.HandleNotifyWX", paynotify.HandleWechatNotify)

	eng.PRequireLogin("forest.ListAnnouncement", ListAnnouncement)

	eng.PRequireLogin("forest.ListMessage", ListMessage)
	eng.PRequireLogin("forest.SetMessageStatus", SetMessageStatus)
	eng.PRequireLogin("forest.DeleteMessages", DeleteMessages)
	eng.PRequireLogin("forest.GetMessageCount", GetMessageCount)

	eng.P("forest.MigrateInterface", mds.HasHeaderStr(mds.AuthHeaderKey, mds.AuthValueMigrateInterface), MigrateInterface)

	// graph node
	eng.PRequireLogin("forest.CreateNode", CreateNode)
	eng.PRequireLogin("forest.GetNodeEdges", GetNodeEdges)
	eng.PRequireLogin("forest.EditNode", EditNode)
	eng.PRequireLogin("forest.CreateNodeEdge", CreateNodeEdge)
	eng.PRequireLogin("forest.DeleteNode", DeleteNode)
	eng.PRequireLogin("forest.GetNodeReference", GetNodeReference)
	eng.PRequireLogin("forest.GetGraphEdges", GetGraphEdges)
	eng.PRequireLogin("forest.RenameNode", RenameNode)

	eng.PRequireLogin("forest.CreateTagGroup", CreateTagGroup)
	eng.PRequireLogin("forest.ListTagGroup", ListTagGroup)
	eng.PRequireLogin("forest.ModifyTagGroup", ModifyTagGroup)
	eng.PRequireLogin("forest.DeleteTagGroup", DeleteTagGroup)
	eng.PRequireLogin("forest.CreateResourceTag", CreateTag)
	eng.PRequireLogin("forest.ListResourceTag", ListTag)
	eng.PRequireLogin("forest.ModifyResourceTag", ModifyTag)
	eng.PRequireLogin("forest.DeleteResourceTag", DeleteTag)
	eng.PRequireLogin("forest.SetResourceTag", SetResourceTag)
	eng.PRequireLogin("forest.GetTagTree", GetTagTree)

	// Resource Perm
	eng.PRequireLogin("forest.SetResourcePerm", SetResourcePerm)
	eng.PRequireLogin("forest.GetResourcePerm", GetResourcePerm)

	// 同义词
	eng.PRequireLogin("forest.ListSynonymKeywords", ListSynonymKeywords)
	eng.PRequireLogin("forest.GetSynonymKeyword", GetSynonymKeyword)
	eng.PRequireLogin("forest.CreateSynonymKeyword", CreateSynonymKeyword)
	eng.PRequireLogin("forest.DeleteSynonymKeyword", DeleteSynonymKeyword)
	eng.PRequireLogin("forest.UpdateSynonymKeyword", UpdateSynonymKeyword)

	eng.PRequireLogin("forest.CreateMajorKeyword", CreateMajorKeyword)
	eng.PRequireLogin("forest.DeleteMajorKeyword", DeleteMajorKeyword)
	eng.PRequireLogin("forest.UpdateMajorKeyword", UpdateMajorKeyword)
	eng.PRequireLogin("forest.ListMajorKeywords", ListMajorKeywords)
	eng.PRequireLogin("forest.GetMajorKeyword", GetMajorKeyword)

	// 热词管理
	eng.PRequireLogin("forest.GetHotWords", GetHotWords)
	eng.PRequireLogin("forest.ListLikes", ListLikes)
	eng.PRequireLogin("forest.MarkResourceLike", MarkResourceLike)

	eng.PRequireLogin("forest.ListCollection", ListCollection)
	eng.PRequireLogin("forest.MarkResourceCollection", MarkResourceCollection)

	// 资源权限范围
	eng.PRequireLogin("forest.SetResourceScope", SetResourceScope)
	eng.PRequireLogin("forest.GetResourceScope", GetResourceScope)
}
