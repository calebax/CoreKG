package svrlkx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtolkx"
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/apps/admin/models/lkx"
	"github.com/insmtx/corekg/pkgs/utils/notify/sms"
	"github.com/redis/go-redis/v9"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
)

// SendVerifyCode 发送验证码
func SendVerifyCode(ctx *gin.Context, req *dtolkx.SendVerifyCodeRequest) error {
	err := CustomerSendSms(1, req.Request.Phone)
	if err != nil {
		logs.ErrorContextf(ctx, "send verify code failed, %s", err)
	}
	return err
}

// VerifyCodeAndSave 验证验证码保存信息
func VerifyCodeAndSave(ctx *gin.Context, req *dtolkx.VerifyVerifyCodeRequest) error {
	err := CustomerVerifySms(1, req.Request.Data.Phone, req.Request.VerifyCode)
	if err != nil {
		logs.ErrorContextf(ctx, "verify code and save failed, %s", err)
		return err
	}
	// 如果验证成功保存信息
	data := &admintype.LkxCustomerInfo{
		Phone:       req.Request.Data.Phone,
		Name:        req.Request.Data.Name,
		Email:       req.Request.Data.Email,
		Position:    req.Request.Data.Position,
		CompanyID:   req.Request.Data.CompanyID,
		CompanyName: req.Request.Data.CompanyName,
		Province:    req.Request.Data.Province,
		City:        req.Request.Data.City,
		Industry:    req.Request.Data.Industry,
		Produce:     req.Request.Data.Produce,
		Description: req.Request.Data.Description,
	}
	err = lkx.SaveInfo(ctx, data)
	if err != nil {
		logs.ErrorContextf(ctx, "save info failed, %s", err)
		return err
	}
	// 保存成功发送给留资机器人
	SendWeComMsg(data)
	return err
}

// 创建留资机器人
func SendWeComMsg(data *admintype.LkxCustomerInfo) {
	webhook := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=ecf5cfc5-1e03-4134-9e00-f76437a4bc42"

	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": "**用户留资信息**\n\n" +
				"> 姓名：" + data.Name + "\n" +
				"> 手机号：" + data.Phone + "\n" +
				"> 邮箱：" + data.Email + "\n" +
				"> 职位：" + data.Position + "\n" +
				"> 公司名称：" + data.CompanyName + "\n" +
				"> 省份：" + data.Province + "\n" +
				"> 城市：" + data.City + "\n" +
				"> 行业：" + data.Industry + "\n" +
				"> 产品：" + data.Produce + "\n" +
				"> 描述：" + data.Description + "\n",
		},
	}

	bs, _ := json.Marshal(body)

	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("POST", webhook, bytes.NewBuffer(bs))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return
}

// RdsKeyCustomerSendSms ...
func RdsKeyCustomerSendSms(identify string, phone string) string {
	cacheKey := fmt.Sprintf("dotpen-api:account:CustomerSendSms:Identify:%v:phone:%v", identify, phone)
	return cacheKey
}

// CustomerSendSms 发送验证
func CustomerSendSms(id uint, phone string) error {
	resp := &apiobj.BaseResponse{}
	key := RdsKeyCustomerSendSms(strconv.Itoa(int(id)), phone)
	is, ttl := redispool.IsExistKey(key)
	if is && ttl > time.Minute*4 {
		resp.Code = errcode.ErrCode_SendVerifyCodeTooBusy
		resp.Message = "短信已发出，请不要频繁发送"
		return errors.New("短信发送频繁，请稍后再试")
	}

	code := random.RandString(6, random.NUMBER)
	// if false && config.IsTestEnv() {
	// 	if err := redispool.SetString(key, code, time.Minute*5); err != nil {
	// 		return err
	// 	}
	// 	logs.Errorf("[main] send sms code: %s, to: %s", code, phone)
	// 	return nil
	// }
	err := sms.SendVerifyCode("lkx", "sms_send_verify_code",
		phone, code, "")
	if err != nil {
		logs.Errorf("[main] send sms failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "短信发送失败"
		return err
	}
	logs.Debugf("send sms success, %v, %v", phone, code)
	if err := redispool.SetString(key, code, time.Minute*5); err != nil {
		logs.Errorf("[main] send sms set cache failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "短信发送失败"
		return err
	}
	return nil
}

// CustomerVerifySms 短信验证
func CustomerVerifySms(id uint, phone, verifyCode string) error {
	resp := &apiobj.BaseResponse{}
	// 1.先验证短信
	key := RdsKeyCustomerSendSms(strconv.Itoa(int(id)), phone)
	code, err := redispool.GetString(key)
	if err != nil {
		if err == redis.Nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "验证码不存在"
			return errors.New("验证码不存在")
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "系统错误，请稍后重试"
		return err
	}
	if code != verifyCode {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "验证码错误"
		return errors.New("验证码错误")
	}
	redispool.Del(key)
	return nil
}
