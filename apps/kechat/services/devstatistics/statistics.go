package devstatistics

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtostatistics"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/xuri/excelize/v2"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetAgentQuestionExcel 获取agent问题excel统计文件
func GetAgentQuestionExcel(ctx *gin.Context, req *dtostatistics.GetAgentQuestionExcelRequest) (res *dtostatistics.GetAgentQuestionExcelResponse, err error) {
	res = &dtostatistics.GetAgentQuestionExcelResponse{}
	// 分页查询全部的question
	questions, err := chatquestion.SearchAgentAllHistory(ctx, &req.Request.StatisticsReq)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "查找失败"
		return res, nil
	}
	// 保存excel
	// 创建一个新的 Excel 文件
	ex := excelize.NewFile()
	defer func() {
		if err := ex.Close(); err != nil {
			logs.ErrorContextf(ctx, "GetAgentQuestionExcel failed to close Excel file: %v", err)
		}
	}()
	sheetName := "Sheet1"
	index, err := ex.NewSheet(sheetName)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "生成sheet失败"
		return res, nil
	}
	title := []string{"创建时间", "问题", "回答"}
	for i, v := range title {
		cellName, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return nil, fmt.Errorf("failed to generate cell location: %v", err)
		}
		ex.SetCellValue(sheetName, cellName, v)
	}
	for i, v := range questions.Hits.Hits {
		rowIndex := i + 2
		cellName, err := excelize.CoordinatesToCellName(1, rowIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to generate cell location: %v", err)
		}
		ex.SetCellValue(sheetName, cellName, v.Source.CreatedAt)
		cellName, err = excelize.CoordinatesToCellName(2, rowIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to generate cell location: %v", err)
		}
		ex.SetCellValue(sheetName, cellName, v.Source.Question)
		cellName, err = excelize.CoordinatesToCellName(3, rowIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to generate cell location: %v", err)
		}
		ex.SetCellValue(sheetName, cellName, v.Source.Answer)
	}
	// 设置默认工作表（可选）
	ex.SetActiveSheet(index)
	// ex.SaveAs("test.xlsx")
	exbuffer, err := ex.WriteToBuffer()
	if err != nil {
		logs.ErrorContextf(ctx, "writing Excel data failed: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "生成excel失败"
		return res, nil
	}
	// 设置响应头
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=agent_%d_history.xlsx", req.Request.AgentID))
	ctx.Header("Content-Length", strconv.Itoa(exbuffer.Len()))
	// 返回 buffer 内容
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", exbuffer.Bytes())
	return res, nil
}

// GetAgentQuestionCount 获取agent问题数量
func GetAgentQuestionCount(ctx *gin.Context, req *dtostatistics.GetAgentQuestionCountRequest) (res *dtostatistics.GetAgentQuestionCountResponse, err error) {
	res = &dtostatistics.GetAgentQuestionCountResponse{}
	// 分页查询全部的question
	count, err := chatquestion.GetAgentHistoryCount(ctx, &req.Request.StatisticsReq)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "查找失败"
		return res, nil
	}
	res.Response.Count = count
	return res, nil
}
