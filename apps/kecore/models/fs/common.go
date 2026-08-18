package fs

import (
	"context"
	"net/url"
	"strings"

	"github.com/ygpkg/yg-go/logs"
)

func IsFilePath(path string) bool {
	return !IsDirPath(path)
}
func IsDirPath(path string) bool {
	return path == "" || strings.HasSuffix(path, "/")
}

// ConvertS3Url 私有化场景使用工具函数，根据referer替换目录
func ConvertS3Url(ctx context.Context, url_p, ref string) string {
	url_obj, err := url.Parse(url_p)
	if err != nil {
		logs.ErrorContextf(ctx, "get url_object error: %v", err)
		return url_p
	}
	url_ref, err := url.Parse(ref)
	if err != nil {
		logs.ErrorContextf(ctx, "get url_ref error: %v", err)
		return url_p
	}
	url_obj.Scheme = url_ref.Scheme
	url_obj.Host = url_ref.Host
	return url_obj.String()
}

// SplitHost 裁剪公共路径前缀
func SplitHost(ctx context.Context, URL string) string {
	u, err := url.Parse(URL)
	if err != nil {
		logs.ErrorContextf(ctx, "get url_object error: %v", err)
		return URL
	}
	pathParts := strings.Split(u.Path, "/")
	if len(pathParts) > 1 && pathParts[1] == Cfg.S3.Bucket {
		u.Scheme = ""
		u.Host = ""
	}

	return u.String()
}

func SpliceUrl(ctx context.Context, path, url_ref string) string {
	ref, err := url.Parse(url_ref)
	if err != nil {
		logs.ErrorContextf(ctx, "get url_ref error: %v", err)
		return path
	}
	newUrl := &url.URL{
		Scheme: ref.Scheme,
		Host:   ref.Host,
		Path:   path,
	}
	return newUrl.String()
}
