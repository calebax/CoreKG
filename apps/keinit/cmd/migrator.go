package main

import (
	"context"

	"github.com/insmtx/corekg/apps/keinit/migrator"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"
)

func runMigratorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "migrator",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := logs.WithContextFields(context.Background(), "cmd", "migrator")
			logs.InfoContextf(ctx, "start migrator")
			if err := runMigrator(ctx); err != nil {
				logs.ErrorContextf(ctx, "migrator cmd run failed, %s", err)
				return
			}
		},
	}
	return cmd
}

func runMigrator(ctx context.Context) error {
	if err := migrator.Run(ctx); err != nil {
		logs.ErrorContextf(ctx, "migrator failed, %s", err)
		return err
	}
	logs.InfoContextf(ctx, "migrator success")
	return nil
}
