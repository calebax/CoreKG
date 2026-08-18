package helper

import (
	"context"
	"os"
	"strings"

	"github.com/ygpkg/yg-go/logs"
)

// ReadENV 读取环境变量
func ReadENV(ctx context.Context, envFiles ...string) (map[string]string, error) {
	env := map[string]string{}
	for _, s := range os.Environ() {
		sli := strings.SplitN(s, "=", 2)
		if len(sli) == 2 {
			k := strings.TrimSpace(sli[0])
			v := strings.TrimSpace(sli[1])
			env[k] = v
		}
	}

	for _, envFile := range envFiles {
		envs, err := os.ReadFile(envFile)
		if err != nil {
			logs.ErrorContextf(ctx, "read env file (%s) failure: %v", envFile, err)
			return nil, err
		}
		for _, s := range strings.Split(string(envs), "\n") {
			sli := strings.SplitN(s, "=", 2)
			if len(sli) == 2 {
				k := strings.TrimSpace(sli[0])
				v := strings.TrimSpace(sli[1])
				env[k] = v
			}
		}
	}
	return env, nil
}
