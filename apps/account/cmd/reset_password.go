package main

import (
	"time"

	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"golang.org/x/crypto/bcrypt"
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

			users, err := user.List(cmd.Context(), apiobj.PageQuery{})
			if err != nil {
				logs.ErrorContextf(cmd.Context(), "[reset-password] list users failed, %s", err)
				return
			}
			for _, user := range users {
				if *user.Password != "" {
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
				user.Password = types.String(string(types.Password(encPassword)))
				if err := db.WithContext(cmd.Context()).Save(&user).Error; err != nil {
					logs.ErrorContextf(cmd.Context(), "[reset-password][%s] save user failed, %s", user.Name, err)
					return
				}
				logs.InfoContextf(cmd.Context(), "[reset-password][%s] password reset to %s", user.Name, passwd)
			}
		},
	}

	return cmd
}

// func addUser() {
// 	empitem := employee.CreateEmployeeItem{
// 		CompanyID: 1,
// 		Username:  "zhangshuyu",
// 		Email:     "",
// 		Mobile:    "",
// 		RealName:  "张舒俞",
// 	}
// 	emp, err := employee.CreateEmployee(dbutil.Account(), empitem)
// 	if err != nil {
// 		logs.Errorf("presetDatabase CreateEmployee err: %v", err)
// 	}
// 	emp.UnionID = types.String("o0hQk6aQsVWa6reHkAUA0utwIFuk")

// 	// emprel := accounttype.EmployeeThirdBinding{
// 	// 	EmployeeID: emp.ID,
// 	// 	BindType:   accounttype.BindTypeWechat,
// 	// 	BindValue:  "o0hQk6aQsVWa6reHkAUA0utwIFuk",
// 	// }
// 	dbutil.Account().Save(&emp)

// }
