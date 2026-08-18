package httptools

import "net/http"

var (
	mimeTypeMap = map[string]string{
		"text/html":                     ".html",
		"text/css":                      ".css",
		"application/javascript":        ".js",
		"application/json":              ".json",
		"image/jpeg":                    ".jpg",
		"image/png":                     ".png",
		"image/gif":                     ".gif",
		"image/svg+xml":                 ".svg",
		"image/x-icon":                  ".ico",
		"image/webp":                    ".webp",
		"image/bmp":                     ".bmp",
		"image/tiff":                    ".tiff",
		"image/x-tiff":                  ".tiff",
		"image/x-ms-bmp":                ".bmp",
		"application/font-woff":         ".woff",
		"application/font-woff2":        ".woff2",
		"application/font-ttf":          ".ttf",
		"application/font-otf":          ".otf",
		"application/vnd.ms-fontobject": ".eot",
		"audio/mpeg":                    ".mp3",
		"video/mp4":                     ".mp4",
		"video/x-m4v":                   ".m4v",
		"video/quicktime":               ".mov",
		"video/webm":                    ".webm",
		"video/x-flv":                   ".flv",
		"application/x-shockwave-flash": ".swf",
		"application/pdf":               ".pdf",
		"application/msword":            ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"application/vnd.ms-excel": ".xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
		"application/vnd.ms-powerpoint":                                             ".ppt",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
		"text/plain":                   ".txt",
		"application/rtf":              ".rtf",
		"text/xml":                     ".xml",
		"application/zip":              ".zip",
		"application/x-rar-compressed": ".rar",
		"application/x-7z-compressed":  ".7z",
		"application/x-bzip2":          ".bz2",
		"application/x-gzip":           ".gz",
		"application/x-xz":             ".xz",
		"application/x-tar":            ".tar",
	}
)

// TransformContentType2Ext
func TransformContentType2Ext(hdr http.Header) string {
	ct := hdr.Get("Content-Type")
	if ct == "" {
		return ""
	}

	ext, ok := mimeTypeMap[ct]
	if !ok {
		return ""
	}
	return ext
}
