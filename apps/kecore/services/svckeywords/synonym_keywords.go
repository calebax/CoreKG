package svckeywords

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtokeywords"
	"github.com/insmtx/corekg/apps/kecore/models/forestkeywords"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// ListSynonymKeywords 获取同义词列表
func ListSynonymKeywords(ctx *gin.Context, req *dtokeywords.ListSynonymKeywordsRequest) (res *dtokeywords.ListSynonymKeywordsResponse, err error) {
	res = &dtokeywords.ListSynonymKeywordsResponse{}
	dao := forestkeywords.NewKeywordsDao()

	data, count, err := dao.ListSynonymKeywords(ctx, &forestkeywords.KeywordsCond{
		BaseCond: forestkeywords.BaseCond{
			CompanyID: runtime.CompanyID(ctx),
			Limit:     req.Request.Limit,
			Offset:    req.Request.Offset,
		},
		LikeWord: req.Request.Word,
	})
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取同义词列表失败"
		return res, nil
	}
	res.Response.Data = data
	res.Response.Total = count
	res.Response.Limit = req.Request.Limit
	res.Response.Offset = req.Request.Offset
	return res, nil
}

// GetSynonymKeyword 获取同义词详情
func GetSynonymKeyword(ctx *gin.Context, req *dtokeywords.GetSynonymKeywordRequest) (res *dtokeywords.GetSynonymKeywordResponse, err error) {
	res = &dtokeywords.GetSynonymKeywordResponse{}
	dao := forestkeywords.NewKeywordsDao()

	data, err := dao.GetSynonymKeywords(ctx, req.Request.ID)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取同义词失败"
		return res, nil
	}
	res.Response.Data = *data
	return res, nil
}

// CreateSynonymKeyword 创建同义词
func CreateSynonymKeyword(ctx *gin.Context, req *dtokeywords.CreateSynonymKeywordRequest) (res *dtokeywords.CreateSynonymKeywordResponse, err error) {
	res = &dtokeywords.CreateSynonymKeywordResponse{}
	// 校验是否存在
	dao := forestkeywords.NewKeywordsDao()
	count, err := dao.CountByCond(ctx, &forestkeywords.KeywordsCond{
		BaseCond: forestkeywords.BaseCond{
			CompanyID: runtime.CompanyID(ctx),
		},
		SubjectID: -1,
		WordType:  foresttype.WordTypeSynonym,
		Words:     append(req.Request.ChildWords, req.Request.Word),
	})
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "重复校验失败"
		return res, nil
	}
	if count != 0 {
		res.Code = errcode.ErrCode_InternalError
		res.Message = fmt.Sprintf("有%d个词已经存在", count)
		return res, nil
	}
	err = dao.DB(ctx).Transaction(func(tx *gorm.DB) error {
		parent := &foresttype.Keywords{
			CompanyID: runtime.CompanyID(ctx),
			Uin:       runtime.Uin(ctx),
			Word:      req.Request.Word,
			WordType:  foresttype.WordTypeSynonym,
		}
		err := dao.WithTx(tx).Insert(ctx, parent)
		if err != nil {
			logs.ErrorContextf(ctx, "Insert parent fail, err: %v", err)
			return err
		}
		childs := []foresttype.Keywords{}
		for _, v := range req.Request.ChildWords {
			childs = append(childs, foresttype.Keywords{
				CompanyID: parent.CompanyID,
				Uin:       parent.Uin,
				Word:      v,
				WordType:  foresttype.WordTypeSynonym,
				SubjectID: parent.ID,
			})
		}
		err = dao.WithTx(tx).BatchInsert(ctx, childs)
		if err != nil {
			logs.ErrorContextf(ctx, "BatchInsert childs fail, err: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		logs.ErrorContextf(ctx, "CreateSynonymKeyword Transaction insert err:%v", err)
		res.Message = "创建失败"
		return res, nil
	}

	return res, nil
}

// DeleteSynonymKeyword 删除同义词关键词
func DeleteSynonymKeyword(ctx *gin.Context, req *dtokeywords.DeleteSynonymKeywordRequest) (res *dtokeywords.DeleteSynonymKeywordResponse, err error) {
	res = &dtokeywords.DeleteSynonymKeywordResponse{}
	dao := forestkeywords.NewKeywordsDao()
	err = dao.DB(ctx).Transaction(func(tx *gorm.DB) error {
		err = dao.WithTx(tx).DeleteBySubjectID(ctx, req.Request.ID, true)
		if err != nil {
			logs.ErrorContextf(ctx, "Delete synonym keyword fail, err: %v", err)
			return err
		}
		err := dao.WithTx(tx).Delete(ctx, req.Request.ID)
		if err != nil {
			logs.ErrorContextf(ctx, "Delete synonym keyword fail, err: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "删除关键词失败"
		return res, nil
	}
	return res, nil
}

// UpdateSynonymKeyword 修改同义词内容
func UpdateSynonymKeyword(ctx *gin.Context, req *dtokeywords.UpdateSynonymKeywordRequest) (res *dtokeywords.UpdateSynonymKeywordResponse, err error) {
	res = &dtokeywords.UpdateSynonymKeywordResponse{}
	// 删
	// 改
	// 增
	dao := forestkeywords.NewKeywordsDao()
	keyword, err := dao.GetByID(ctx, req.Request.ID)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取同义词失败"
		return res, nil
	}
	err = dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err = dao.WithTx(tx).DeleteBySubjectID(ctx, req.Request.ID, false)
		if err != nil {
			logs.ErrorContextf(ctx, "Delete synonym keyword fail, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "删除同义词失败"
			return err
		}
		existWords := req.Request.ChildWords
		if keyword.Word != req.Request.Word {
			existWords = append(existWords, req.Request.Word)
		}
		exist, err := wordsSyIsExist(ctx, dao.WithTx(tx), existWords)
		if err != nil {
			logs.ErrorContextf(ctx, "Delete synonym keyword fail, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "重复校验失败"
			return err
		}
		if exist {
			res.Code = errcode.ErrCode_InternalError
			res.Message = "有重复的词，请检查后重试"
			return fmt.Errorf("has exist words")
		}
		if keyword.Word != req.Request.Word {
			keyword.Word = req.Request.Word
			err = tx.Save(keyword).Error
			if err != nil {
				logs.ErrorContextf(ctx, "Update synonym keyword fail, err: %v", err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "修改主词失败"
				return err
			}
		}
		childs := []foresttype.Keywords{}
		for _, v := range req.Request.ChildWords {
			childs = append(childs, foresttype.Keywords{
				CompanyID: keyword.CompanyID,
				Uin:       keyword.Uin,
				Word:      v,
				WordType:  foresttype.WordTypeSynonym,
				SubjectID: keyword.ID,
			})
		}
		err = dao.WithTx(tx).BatchInsert(ctx, childs)
		if err != nil {
			logs.ErrorContextf(ctx, "BatchInsert childs fail, err: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return res, nil
	}

	return res, nil
}

func wordsSyIsExist(ctx *gin.Context, dao *forestkeywords.KeywordsDao, words []string) (bool, error) {
	count, err := dao.CountByCond(ctx, &forestkeywords.KeywordsCond{
		BaseCond: forestkeywords.BaseCond{
			CompanyID: runtime.CompanyID(ctx),
		},
		WordType:  foresttype.WordTypeSynonym,
		Words:     words,
		SubjectID: -1,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "CountByCond fail, err: %v", err)
		return true, err
	}
	if count != 0 {
		return true, nil
	}
	return false, nil
}
