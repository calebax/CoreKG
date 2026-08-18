package testutils

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/redis/go-redis/v9"
	uuid "github.com/satori/go.uuid"
	"github.com/ygpkg/yg-go/apis/constants"
	runtimeauth "github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"go.uber.org/zap/zapcore"
)

func init() {
	gin.SetMode(gin.TestMode)
	logs.SetLevel(zapcore.DebugLevel)
}

var initializer Initializer
var once sync.Once

// Initialize 初始化指定的应用
func Initialize(appName AppName) {
	if _, ok := appNameMap[appName]; !ok {
		panic(fmt.Sprintf("unsupported app: %s", appName))
	}
	once.Do(func() {
		initializerFunc, ok := initializerMap[appName]
		if !ok {
			panic(fmt.Sprintf("unsupported app: %s", appName))
		}
		inst, err := initializerFunc()
		if err != nil {
			panic(fmt.Sprintf("init %s: %v", appName, err))
		}
		initializer = inst
	})

	if err := initializer.Initialize(); err != nil {
		fmt.Printf("init %s error: %v", appName, err)
	}
}

func Close() {
	if initializer == nil {
		fmt.Println("no initializer")
		return
	}
	if err := initializer.Close(); err != nil {
		fmt.Printf("close %s error: %v", initializer, err)
	}
}

type Initializer interface {
	Initialize() error
	Close() error
}

type InitializerFunc func() (Initializer, error)

var initializerMap = map[AppName]InitializerFunc{
	AppNameKecore:   newKecoreInitializer,
	AppNameKechat:   newKechatInitializer,
	AppNameKesale:   newkesaleInitializer,
	AppNameAdmin:    newAdminInitializer,
	AppNameKesearch: newKesearchInitializer,
}

type Option func(ctx *gin.Context)

func NewCtx(opts ...Option) *gin.Context {
	ctx := &gin.Context{Request: &http.Request{URL: new(url.URL)}}
	ctx.Request = ctx.Request.WithContext(context.Background())
	for _, opt := range opts {
		opt(ctx)
	}
	reqID := hex.EncodeToString(uuid.Must(uuid.NewV4(), nil).Bytes())
	ctx.Set(constants.CtxKeyRequestID, reqID)
	return ctx
}

func WithUin(uin uint) Option {
	return func(ctx *gin.Context) {
		// 生成 token
		var token string
		redisKey := fmt.Sprintf("roc_unit_test_client_token:%d", uin)
		cacheToken, getCacheTokenErr := redispool.Redis().Get(ctx, redisKey).Result()
		if getCacheTokenErr != nil && getCacheTokenErr != redis.Nil {
			panic(fmt.Sprintf("get cache token failed, err: %v", getCacheTokenErr))
		}
		if cacheToken != "" {
			token = cacheToken
		} else {
			token = user.GenerateJwtToken(ctx, uin, runtimeauth.LoginWayUnknown, "127.0.0.1", "yygu")
			if _, err := redispool.Redis().Set(ctx, redisKey, token, time.Hour*24*3).Result(); err != nil {
				panic(fmt.Sprintf("set cache token failed, err: %v", err))
			}
		}
		token = strings.TrimPrefix(token, runtimeauth.AuthBearer)
		token = strings.TrimSpace(token)
		claims := new(runtimeauth.UserClaims)
		_, parseClaimsErr := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if token.Claims == nil {
				return nil, fmt.Errorf("token claims is nil")
			}
			c, ok := token.Claims.(*runtimeauth.UserClaims)
			if !ok {
				return nil, fmt.Errorf("token claims is not UserClaims")
			}

			return runtimeauth.GetJwtSecret(c.Issuer)
		})
		if parseClaimsErr != nil {
			panic(fmt.Sprintf("parse claims failed, err: %v", parseClaimsErr))
		}
		loginStatus := &runtimeauth.LoginStatus{
			Token: token,
			State: runtimeauth.StateSucc,
			Claim: claims,
		}
		user, err := user.GetUserIdentificationByUIN(ctx, uin)
		if err != nil {
			panic(fmt.Sprintf("get user by uin failed, err: %v", err))
		}

		switch user.SubjectType {
		case global.SubjectTypeCompany:
			loginStatus.SetID(global.CtxKeyCompanyID, user.SubjectID)
			em, err := employee.GetEmployeeByUin(user.ID)
			if err != nil {
				panic(fmt.Sprintf("get employee by uin failed, err: %v", err))
			}
			loginStatus.SetID(global.CtxKeyEmployeeID, em.ID)
		}

		loginStatus.SetID(constants.CtxKeyUin, uin)
		ctx.Set(constants.CtxKeyLoginStatus, loginStatus)
	}
}

// OpenFile 打开测试数据文件，relPath 为相对路径，相对于当前文件所在目录
func OpenFile(relPath string) (*os.File, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("cannot get caller info")
	}
	return os.Open(filepath.Join(filepath.Dir(filename), relPath))
}

// TestFilePath 获取测试数据文件路径，relPath 为相对路径，相对于当前文件所在目录
func TestFilePath(relPath string) string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("cannot get caller info")
	}
	return filepath.Join(filepath.Dir(filename), relPath)
}
