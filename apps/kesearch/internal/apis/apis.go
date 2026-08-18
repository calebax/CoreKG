package apis

import (
	"github.com/insmtx/corekg/apps/account/accountmds"
	"github.com/insmtx/corekg/apps/kecore/services/deployhandle"
	"github.com/insmtx/corekg/apps/kesearch/internal/apis/chunkctl"
	"github.com/insmtx/corekg/apps/kesearch/internal/apis/coze"
	"github.com/insmtx/corekg/apps/kesearch/internal/apis/globalsearchctl"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/metrics"
)

// RegistryRouter 注册路由
func RegistryRouter(eng *server.Router) {
	eng.HandleDoc("kesearch")
	eng.G("kesearch.metrics", metrics.GetHandler())
	eng.AuthInject("yygu", accountmds.InjectLoginStatus)

	// 知识库检索
	eng.PRequireLogin("kesearch.ForestSearch", globalsearchctl.ForestSearch)
	eng.PRequireLogin("kesearch.ForestSearchDoc", globalsearchctl.ForestSearchDoc)
	eng.PRequireLogin("kesearch.ForestSearchImage", globalsearchctl.ForestSearchImage)
	eng.PRequireLogin("kesearch.ForestSearchVideo", globalsearchctl.ForestSearchVideo)

	// 全局搜索
	eng.PRequireLogin("kesearch.GlobalSearch", globalsearchctl.GlobalSearch)
	eng.PRequireLogin("kesearch.GlobalSearchDoc", globalsearchctl.GlobalSearchDoc)
	eng.PRequireLogin("kesearch.GlobalSearchImage", globalsearchctl.GlobalSearchImage)
	eng.PRequireLogin("kesearch.GlobalSearchVideo", globalsearchctl.GlobalSearchVideo)
	eng.PRequireLogin("kesearch.GlobalSearchAgent", globalsearchctl.GlobalSearchAgent)
	eng.PRequireLogin("kesearch.GlobalSearchForest", globalsearchctl.GlobalSearchForest)
	eng.PRequireLogin("kesearch.GlobalSearchExternalData", globalsearchctl.GlobalSearchExternalData)

	eng.PRequireLogin("kesearch.ListFileChunk", chunkctl.ListFileChunk)
	eng.PRequireLogin("kesearch.GetChunkByID", chunkctl.GetChunkByID)
	eng.PRequireLogin("kesearch.GetChunkBySequence", GetChunkBySequence)
	eng.PRequireLogin("kesearch.GetChunkDetail", GetChunkDetail)
	eng.PRequireLogin("kesearch.UpdateChunk", chunkctl.UpdateChunk)
	eng.PRequireLogin("kesearch.DeleteChunk", chunkctl.DeleteChunk)
	eng.PRequireLogin("kesearch.DisableFileChunk", chunkctl.DisableFileChunk)

	eng.P("kesearch.MigrateChunkFileName", chunkctl.MigrateChunkFileName)

	eng.P("kesearch.KnowledgeSearch", coze.KnowledgeSearch)

	eng.PRequireLogin("kesearch.RerankSearchChunk", RerankSearchChunk)

	eng.P("kesearch.SwitchPrivateEvn", deployhandle.SwitchPrivateEvn)
	eng.P("kesearch.NowDeployMode", deployhandle.NowDeployMode)

	eng.P("kesearch.ExcuteSql", ExcuteSql)
}
