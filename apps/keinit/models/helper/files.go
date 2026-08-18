package helper

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/ygpkg/yg-go/logs"
)

// ListFileWithExt 返回指定目录下指定后缀的文件
func ListFileWithExt(ctx context.Context, sqlDir string, ext string) ([]string, error) {
	var files []string

	err := filepath.Walk(sqlDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logs.ErrorContextf(ctx, "walk dir error: %v", err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "walk dir error: %v", err)
		return nil, err
	}
	sort.Sort(VersionSlice(files))
	return files, err
}

type VersionSlice []string

func (s VersionSlice) Len() int { return len(s) }

func (s VersionSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s VersionSlice) Less(i, j int) bool {
	return compareVersions(s[i], s[j])
}

func compareVersions(v1, v2 string) bool {
	// 提取版本号部分 (vX.Y_Z 或 vX.Y.Z)
	re := regexp.MustCompile(`v(\d+)\.(\d+)(?:\.(\d+)|_(\d+))`)

	matches1 := re.FindStringSubmatch(v1)
	matches2 := re.FindStringSubmatch(v2)

	if len(matches1) == 0 || len(matches2) == 0 {
		return v1 < v2 // 如果格式不匹配，回退到字典序
	}

	// 提取版本号
	major1, _ := strconv.Atoi(matches1[1])
	minor1, _ := strconv.Atoi(matches1[2])
	patch1 := 0
	if matches1[3] != "" {
		patch1, _ = strconv.Atoi(matches1[3])
	} else if matches1[4] != "" {
		patch1, _ = strconv.Atoi(matches1[4])
	}

	major2, _ := strconv.Atoi(matches2[1])
	minor2, _ := strconv.Atoi(matches2[2])
	patch2 := 0
	if matches2[3] != "" {
		patch2, _ = strconv.Atoi(matches2[3])
	} else if matches2[4] != "" {
		patch2, _ = strconv.Atoi(matches2[4])
	}

	// 比较主版本号
	if major1 != major2 {
		return major1 < major2
	}

	// 比较次版本号
	if minor1 != minor2 {
		return minor1 < minor2
	}

	// 比较修订版本号
	return patch1 < patch2
}
