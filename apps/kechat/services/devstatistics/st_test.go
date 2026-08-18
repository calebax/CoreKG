package devstatistics

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtostatistics"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestGetAgentQuestionExcel(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	chatquestion.InitHistoryESClient(context.Background())
	GetAgentQuestionExcel(&gin.Context{}, &dtostatistics.GetAgentQuestionExcelRequest{
		Request: dtostatistics.GetAgentQuestionExcelEmbedRequest{
			StatisticsReq: chatquestion.StatisticsReq{
				AgentID: 732,
			},
		},
	})
}
