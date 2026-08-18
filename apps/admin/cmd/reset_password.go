package main

import (
	"context"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/admin/models/employee"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	rootCmd.AddCommand(resetPasswordCmd())
}

func resetPasswordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reset-password",
		Aliases: []string{"rp"},
		Short:   "reset password",
		Run: func(cmd *cobra.Command, args []string) {
			logs.DebugContextf(cmd.Context(), "try to reset password")

			cfg := config.Conf()
			db, err := initDatabase(cmd.Context(), cfg, true)
			if err != nil {
				logs.ErrorContextf(cmd.Context(), "[reset-password] init database failed, %s", err)
				return
			}

			users, err := employee.ListEmployees(apiobj.PageQuery{})
			if err != nil {
				logs.ErrorContextf(cmd.Context(), "[reset-password] list users failed, %s", err)
				return
			}
			for _, user := range users {
				if user.Password != "" {
					continue
				}
				if user.Email == nil {
					continue
				}
				passwd := *user.Email + "ABC"
				encPassword, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
				if err != nil {
					logs.ErrorContextf(cmd.Context(), "[reset-password] generate password(%s) failed, %s", passwd, err)
					return
				}
				user.CreatedAt = time.Now()
				user.Password = types.Password(encPassword)
				if err := db.Save(&user).Error; err != nil {
					logs.ErrorContextf(cmd.Context(), "[reset-password][%s] save user failed, %s", user.RealName, err)
					return
				}
				logs.InfoContextf(cmd.Context(), "[reset-password][%s] password reset to %s", user.RealName, passwd)
			}
		},
	}

	return cmd
}

func migrateEmployeeUser(ctx context.Context, db *gorm.DB) {
	db = db.WithContext(ctx)
	empList, err := employee.ListAllEmployees()
	if err != nil {
		logs.ErrorContext(ctx, "[migrate-employee] list all employees failed, %s", err)
		return
	}

	for _, emp := range empList {
		if emp.Uin > 0 {
			continue
		}

		usr, err := user.GetUserByWechatUnionID(*emp.UnionID)
		if err != nil {
			logs.ErrorContextf(ctx, "[migrate-employee] get user by wechat union id(%s) failed, %s", *emp.UnionID, err)
			usr = &accounttype.User{
				Identify:        emp.Username,
				Name:            emp.NickName,
				Bio:             emp.NickName,
				AvatarURL:       emp.AvatarURL,
				Email:           emp.Email,
				Phone:           emp.Mobile,
				WechatUnionID:   emp.UnionID,
				WechatWebOpenID: emp.WebOpenID,
			}
			if err := db.Save(usr).Error; err != nil {
				logs.ErrorContextf(ctx, "[migrate-employee] save user failed, %s", err)
				continue
			}
		}

		uin := &accounttype.UserIdentification{
			UserID:      usr.ID,
			SubjectType: accounttype.SubjectTypeIndividual,
			UinStatus:   accounttype.UinStatusNormal,
			Issuer:      "yyguadmin",
		}
		if err := db.Save(uin).Error; err != nil {
			logs.ErrorContextf(ctx, "[migrate-employee] save user identification failed, %s", err)
			continue
		}

		emp.Uin = uin.ID
		if err := db.Save(&emp).Error; err != nil {
			logs.ErrorContextf(ctx, "[migrate-employee] save employee failed, %s", err)
			continue
		}
		logs.InfoContextf(ctx, "[migrate-employee] employee(%s) uin(%d) migrated", emp.RealName, uin.ID)
	}

}
