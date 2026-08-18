package foresttype

import (
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

const (
	// TableNamePrefix 表名前缀
	TableNamePrefix = "ke_"

	TableNameKnownowParseHistory      = TableNamePrefix + "parse_history"
	TableNameKnownowFileParse         = TableNamePrefix + "file_parse"
	TableNameKnownowFileQA            = TableNamePrefix + "file_qa"
	TableNameKnownowTask              = TableNamePrefix + "task"
	TableNameKnownowForest            = TableNamePrefix + "forest"
	TableNameKnownowForestFile        = TableNamePrefix + "forest_file"
	TableNameKnownowQASession         = TableNamePrefix + "qa_session"
	TableNameKnownowForestQA          = TableNamePrefix + "forest_qa"
	TableNameKnownowForestPublicScope = TableNamePrefix + "forest_public_scope"
	TableNameKeResourceScope          = TableNamePrefix + "resource_scope"
	TableNameKeForestDBInstance       = TableNamePrefix + "forest_db_instance"
	TableNameKeForestDB               = TableNamePrefix + "forest_db"
	TableNameKeForestTable            = TableNamePrefix + "forest_table"
	TableNameKeForestExcelSheet       = TableNamePrefix + "forest_excel_sheet"
	TableNameKeCompanyDb              = TableNamePrefix + "company_db"

	TableNameKeForestGraph        = TableNamePrefix + "forest_graph"
	TableNameKeForestGraphVersion = TableNamePrefix + "forest_graph_version"
	TableNameKeGraphNode          = TableNamePrefix + "graph_node"
	TableNameKeGraphTagNode       = TableNamePrefix + "graph_tag_node"
	TableNameKeGraphTag           = TableNamePrefix + "graph_tag"
	TableNameKeGraphEdgeTag       = TableNamePrefix + "graph_edge_tag"
	TableNameKeGraphEdge          = TableNamePrefix + "graph_edge"

	TableNameKeProject = TableNamePrefix + "project"

	TableNameKeArticle            = TableNamePrefix + "article"
	TableNameKeArticleTemplate    = TableNamePrefix + "article_template"
	TableNameKeArticleHistory     = TableNamePrefix + "article_history"
	TableNameKeGraphResourceChunk = TableNamePrefix + "graph_resource_chunk"
	TableNameKePackage            = TableNamePrefix + "package"
	TableNameKeCompanyQuota       = TableNamePrefix + "company_quota"
	TableNameKeMessageTemplate    = TableNamePrefix + "message_template"
	TableNameKeUinMessage         = TableNamePrefix + "uin_message"
	TableNameTagGroup             = TableNamePrefix + "tag_group"
	TableNameTag                  = TableNamePrefix + "tag"
	TableNameResourceTag          = TableNamePrefix + "resource_tag"
	TableNameRecentUsedTag        = TableNamePrefix + "recent_used_tag"
	TableNameKeywords             = TableNamePrefix + "keywords"
	TableNameForestHotWord        = TableNamePrefix + "forest_hot_word"
	TableNameUinLikes             = TableNamePrefix + "uin_likes"
	TableNameUinCollections       = TableNamePrefix + "uin_collections"
)

// InitDB init db tables
func InitDB() error {
	err := dbtools.InitModel(dbutil.Knownow(),
		&KnownowParseHistory{},
		&KnownowFileQA{},
		&KnownowTask{},

		&KnownowForest{},
		&KnownowForestFile{},
		&KnownowFileParse{},
		&KnownowForestPublicScope{},
		&KeResourceScope{},

		&KnownowQASession{},
		&KnownowForestQA{},

		&ForestDBInstance{},
		&ForestDB{},
		&ForestTable{},
		&ForestExcelSheet{},
		&KeCompanyDB{},

		&ForestGraphVersion{},
		// &GraphNode{},
		&GraphTagNode{},
		&GraphTag{},
		&GraphEdge{},
		&GraphEdgeTag{},

		&KnownowProject{},

		&KeArticle{},
		&KeArticleTemplate{},
		&KeArticleHistory{},
	)

	if err != nil {
		// logs.Errorf( "[main] init foresttype database failed, %s", err)
		return err
	}
	// logs.Infof( "[main] init foresttype database success")

	{
		// db_pre
		if err := presetDatabase(); err != nil {
			return err
		}
	}
	return nil
}

// 数据库初始化准备
func presetDatabase() error {
	return nil
}
