package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/insmtx/corekg/pkgs/utils/notify/sms"
	"github.com/redis/go-redis/v9"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
)

const (
	SMSCodeLen = 6
	// UpdatePhonePosition 更新手机号场景
	UpdatePhonePosition = "UpdatePhone"
	// ContactFormPosition  发起联系表单场景
	ContactFormPosition = "ContactForm"
	// UpgradeFormPosition  发起版本升级表单场景
	UpgradeFormPosition = "UpgradeForm"
	// DotpenContactFormPosition  点笔发起联系表单场景
	DotpenContactFormPosition = "DotpenContactForm"
)

// RdsKeyCustomerSendSms ...
func RdsKeyCustomerSendSms(identify string, phone string) string {
	cacheKey := fmt.Sprintf("roc:account:CustomerSendSms:Identify:%v:phone:%v", identify, phone)
	return cacheKey
}

// CustomerSendSms 发送验证
func CustomerSendSms(ctx context.Context, id uint, phone, position string, resp *apiobj.BaseResponse) error {
	key := RdsKeySendSmsWithPosition(position, strconv.Itoa(int(id)), phone)
	is, ttl := redispool.IsExistKey(key)
	if is && ttl > time.Minute*2 {
		resp.Code = errcode.ErrCode_SendVerifyCodeTooBusy
		resp.Message = "account_sms_send_too_frequent" // 短信已发出，请不要频繁发送
		return nil
	}

	code := random.RandString(SMSCodeLen, random.NUMBER)
	// if false && config.IsTestEnv() {
	// 	if err := redispool.SetString(key, code, time.Minute*5); err != nil {
	// 		return err
	// 	}
	// 	logs.Errorf("[main] send sms code: %s, to: %s", code, phone)
	// 	return nil
	// }
	err := sms.SendVerifyCode("account", "sms_send_verify_code",
		phone, code, "")
	if err != nil {
		logs.ErrorContextf(ctx, "[main] send sms failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_sms_send_failed" // 短信发送失败
		return err
	}
	logs.DebugContextf(ctx, "send sms success, %v, %v", phone, code)
	if err := redispool.SetString(key, code, time.Minute*5); err != nil {
		logs.ErrorContextf(ctx, "[main] send sms set cache failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_sms_send_failed" // 短信发送失败
		return err
	}
	return nil
}

// CustomerVerifySms 短信验证
func CustomerVerifySms(id uint, phone, position, verifyCode string, resp *apiobj.BaseResponse) error {
	// 1.先验证短信
	key := RdsKeySendSmsWithPosition(position, strconv.Itoa(int(id)), phone)
	code, err := redispool.GetString(key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_sms_code_not_exist" // 验证码不存在
			return errors.New("sms code not exist")
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_system_error" // 系统错误，请稍后重试
		return err
	}
	if code != verifyCode {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_sms_code_invalid" // 验证码错误
		return errors.New("sms code invalid")
	}
	redispool.Del(key)
	return nil
}

func RdsKeySendSmsWithPosition(position, identify string, phone string) string {
	if len(position) == 0 {
		position = "default"
	}
	cacheKey := fmt.Sprintf("roc:account:CustomerSendSms:Position:%v,Identify:%v:phone:%v", position, identify, phone)
	return cacheKey
}
