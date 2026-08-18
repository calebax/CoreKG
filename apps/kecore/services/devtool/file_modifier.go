package devtool

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ygpkg/yg-go/logs"
)

type ModifierArgs map[string]string

// FileModifier 文件修改函数类型，接收本地文件路径，返回修改后的内容、路径、是否实际修改了内容、错误
type FileModifier func(ctx context.Context, localPath string, content []byte, args ModifierArgs) (newContent []byte, newPath string, modified bool, err error)

func ReplaceURLModifier(ctx context.Context) FileModifier {
	return func(_ context.Context, localPath string, content []byte, args ModifierArgs) ([]byte, string, bool, error) {

		oldPrefix := args["oldPrefix"]
		newPrefix := args["newPrefix"]
		bizBucket := args["bizBucket"]
		service := args["service"]

		if oldPrefix == "" || newPrefix == "" {
			return content, localPath, false, fmt.Errorf("missing oldPrefix or newPrefix")
		}

		// 构建动态 pattern
		dynamicPath := "/" + bizBucket + "/" + service + "/"
		logs.InfoContextf(ctx, "[ReplaceURLModifier] dynamicPath: %s", dynamicPath)
		// e.g. http://ip:port/test-bucket/servicea/xxxxx
		pattern := regexp.MustCompile(
			regexp.QuoteMeta(oldPrefix) + regexp.QuoteMeta(dynamicPath) + `[^\s\)]*`,
		)

		matches := pattern.FindAllString(string(content), -1)
		if len(matches) == 0 {
			logs.InfoContextf(ctx, "[ReplaceURLModifier] no matching URLs found for oldPrefix: %s", oldPrefix)
			return content, localPath, false, nil
		}

		logs.InfoContextf(ctx, "找到 %d 个需要替换的 URL", len(matches))

		replaced := pattern.ReplaceAllStringFunc(string(content), func(match string) string {
			return newPrefix + match[len(oldPrefix):]
		})

		return []byte(replaced), localPath, true, nil
	}
}
