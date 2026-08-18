package main

import (
	"context"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/keinit/models/es"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"
)

func initESCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "init-es",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := logs.WithContextFields(context.Background(), "cmd", "init-es")
			logs.InfoContextf(ctx, "start init-es")
			if err := initES(ctx); err != nil {
				logs.ErrorContext(ctx, "init-es failed, %s", err)
				return
			}
			logs.InfoContextf(ctx, "end init-es")
		},
	}
	return cmd
}

func initES(ctx context.Context) error {
	var (
		escli *elasticsearch.Client
		err   error
	)

	for {
		escli, err = essearch.InitESClient(ctx)
		if err != nil {
			logs.WarnContextf(ctx, "NewWrapper InitESClient error: %v, wait 10s and retry", err)
			time.Sleep(time.Second * 10)
			continue
		}

		resp, err := escli.Info()
		if err != nil {
			logs.ErrorContextf(ctx, "es query failed: %v, wait 10s and retry", err)
			time.Sleep(time.Second * 10)
			continue
		}
		logs.DebugContextf(ctx, "es query success: %v", resp.String())
		break
	}
	err = es.ExecuteDSLWithFile(ctx, escli, esDSLDir)
	if err != nil {
		logs.ErrorContextf(ctx, "es execute dsl failed: %v", err)
		return err
	}

	return nil
}
