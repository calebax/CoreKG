package svckeywords

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtokeywords"
	"github.com/insmtx/corekg/apps/kecore/models/forestkeywords"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// CreateMajorKeyword 创建专业术语
func CreateMajorKeyword(ctx *gin.Context, req *dtokeywords.CreateMajorKeywordRequest) (res *dtokeywords.CreateMajorKeywordResponse, err error) {
	res = &dtokeywords.CreateMajorKeywordResponse{}
	dao := forestkeywords.NewKeywordsDao()
	count, err := dao.CountByCond(ctx, &forestkeywords.KeywordsCond{
		BaseCond: forestkeywords.BaseCond{
			CompanyID: runtime.CompanyID(ctx),
		},
		WordType: foresttype.WordTypeMajor,
		Words:    []string{req.Request.Word},
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
	m := &foresttype.Keywords{
		CompanyID:   runtime.CompanyID(ctx),
		Uin:         runtime.Uin(ctx),
		Word:        req.Request.Word,
		Description: req.Request.Description,
		WordType:    foresttype.WordTypeMajor,
	}
	err = dao.Insert(ctx, m)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "创建专业术语失败"
		return res, nil
	}
	return res, nil
}

// DeleteMajorKeyword 删除专业术语
func DeleteMajorKeyword(ctx *gin.Context, req *dtokeywords.DeleteMajorKeywordRequest) (res *dtokeywords.DeleteMajorKeywordResponse, err error) {
	res = &dtokeywords.DeleteMajorKeywordResponse{}
	dao := forestkeywords.NewKeywordsDao()
	err = dao.Delete(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "Delete synonym keyword fail, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "删除专业术语失败"
		return res, nil
	}
	return res, nil
}

// UpdateMajorKeyword 修改专业术语
func UpdateMajorKeyword(ctx *gin.Context, req *dtokeywords.UpdateMajorKeywordRequest) (res *dtokeywords.UpdateMajorKeywordResponse, err error) {
	res = &dtokeywords.UpdateMajorKeywordResponse{}
	dao := forestkeywords.NewKeywordsDao()
	keyword, err := dao.GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateMajorKeyword GetByID fail, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取专业术语失败"
		return res, nil
	}

	keyword.Description = req.Request.Description
	if keyword.Word != req.Request.Word {
		keyword.Word = req.Request.Word
		exist, err := wordsMajorIsExist(ctx, dao, []string{req.Request.Word})
		if err != nil {
			logs.ErrorContextf(ctx, "wordsMajorIsExist major keyword fail, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "重复校验失败"
			return res, nil
		}
		if exist {
			res.Code = errcode.ErrCode_InternalError
			res.Message = "有重复的词，请检查后重试"
			return res, nil
		}
	}

	err = dao.DB(ctx).Save(keyword).Error
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateMajorKeyword Save fail, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "修改专业术语失败"
		return res, nil
	}

	return res, nil
}

func ListMajorKeywords(ctx *gin.Context, req *dtokeywords.ListMajorKeywordsRequest) (res *dtokeywords.ListMajorKeywordsResponse, err error) {
	res = &dtokeywords.ListMajorKeywordsResponse{}
	dao := forestkeywords.NewKeywordsDao()

	data, count, err := dao.GetPageListByCond(ctx, &forestkeywords.KeywordsCond{
		BaseCond: forestkeywords.BaseCond{
			CompanyID: runtime.CompanyID(ctx),
			Limit:     req.Request.Limit,
			Offset:    req.Request.Offset,
		},
		LikeWord:  req.Request.Word,
		SubjectID: 0,
		WordType:  foresttype.WordTypeMajor,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "ListMajorKeywords GetPageListByCond fail, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取专业术语失败"
		return res, nil
	}
	userMap := map[uint]*accounttype.User{}
	// map组装
	for _, parent := range data {
		userEntity, exists := userMap[parent.Uin]
		if !exists {
			userEntity, err = user.GetUserByUin(ctx, parent.Uin)
			if err != nil {
				logs.ErrorContextf(ctx, "GetUserByUin error: %v", err)
				continue
			}
			userMap[parent.Uin] = userEntity
		}
		res.Response.Data = append(res.Response.Data, &forestkeywords.MajorKeywordDetail{
			Keywords: parent,
			UserName: userEntity.Name,
		})
	}
	res.Response.Total = count
	res.Response.Limit = req.Request.Limit
	res.Response.Offset = req.Request.Offset
	return res, nil
}

// GetMajorKeyword 获取专业术语
func GetMajorKeyword(ctx *gin.Context, req *dtokeywords.GetMajorKeywordRequest) (res *dtokeywords.GetMajorKeywordResponse, err error) {
	res = &dtokeywords.GetMajorKeywordResponse{}
	dao := forestkeywords.NewKeywordsDao()
	keyword, err := dao.GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateMajorKeyword GetByID fail, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取专业术语失败"
		return res, nil
	}

	res.Response.Data = keyword
	return res, nil
}

func wordsMajorIsExist(ctx *gin.Context, dao *forestkeywords.KeywordsDao, words []string) (bool, error) {
	count, err := dao.CountByCond(ctx, &forestkeywords.KeywordsCond{
		BaseCond: forestkeywords.BaseCond{
			CompanyID: runtime.CompanyID(ctx),
		},
		WordType:  foresttype.WordTypeMajor,
		Words:     words,
		SubjectID: 0,
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
