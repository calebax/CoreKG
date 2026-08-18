package utils

import (
	"fmt"
	"time"
)

func CurrentDateInfoRFC1123() string {
	// 获取当前时间（本地时区）
	now := time.Now()

	// 获取时区名与偏移量
	_, offset := now.Zone()
	hours := offset / 3600
	minutes := (offset % 3600) / 60

	// 格式化时区字符串
	sign := "+"
	if hours < 0 || minutes < 0 {
		sign = "-"
		hours = -hours
		minutes = -minutes
	}
	timezone := fmt.Sprintf("GMT%s%02d:%02d", sign, hours, minutes)

	// RFC1123 格式时间
	formatted := now.Format("Mon, 02 Jan 2006 15:04:05")

	return fmt.Sprintf("%s %s", formatted, timezone)
}
