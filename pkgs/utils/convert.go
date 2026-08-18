package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ByteUnit 表示字节单位类型（string 的别名）
type ByteUnit string

// 单位常量（对外可见）
const (
	UnitB  ByteUnit = "B"
	UnitKB ByteUnit = "KB"
	UnitMB ByteUnit = "MB"
	UnitGB ByteUnit = "GB"
	UnitTB ByteUnit = "TB"
	UnitPB ByteUnit = "PB"
	UnitEB ByteUnit = "EB"
)

// 内部换算因子（私有）
const (
	_          = iota
	kb float64 = 1 << (10 * iota) // 1024
	mb
	gb
	tb
	pb
	eb
)

// ConvertBytes 将字节数转为指定单位的 float64 值（保留两位小数）
// unit 不区分大小写，例如 "kb"、"MB" 都可以
func ConvertBytes(n uint, unit ByteUnit) (float64, error) {
	u := strings.ToUpper(strings.TrimSpace(string(unit)))

	var val float64
	switch ByteUnit(u) {
	case UnitB:
		val = float64(n)
	case UnitKB:
		val = float64(n) / kb
	case UnitMB:
		val = float64(n) / mb
	case UnitGB:
		val = float64(n) / gb
	case UnitTB:
		val = float64(n) / tb
	case UnitPB:
		val = float64(n) / pb
	case UnitEB:
		val = float64(n) / eb
	default:
		return 0, fmt.Errorf("unsupported unit: %s", unit)
	}

	// 保留两位小数
	return RoundFloat(val, 2)
}

func RoundFloat(num float64, precision int) (float64, error) {
	precisionRule := fmt.Sprintf("%%.%df", precision)
	return strconv.ParseFloat(fmt.Sprintf(precisionRule, num), 64)
}

// VToUint64 将任意类型转换为 uint64 类型
func VToUint64(v any) uint64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return uint64(n)
	case int64:
		return uint64(n)
	case int8:
		return uint64(n)
	case int32:
		return uint64(n)
	case uint:
		return uint64(n)
	case uint64:
		return n
	case uint8:
		return uint64(n)
	case uint32:
		return uint64(n)
	case float64:
		return uint64(n)
	case float32:
		return uint64(n)
	case string:
		i, _ := strconv.ParseUint(n, 10, 64)
		return i
	}
	return 0
}

func Percentage[T int64 | float64](numerator, denominator T, decimals int) string {
	// 处理分母为0的情况
	if denominator == 0 {
		return "0.00%"
	}

	// 转换为 float64 进行计算
	num := float64(numerator)
	den := float64(denominator)

	// 计算百分比
	percentage := (num / den) * 100

	// 根据小数位数进行四舍五入
	multiplier := math.Pow(10, float64(decimals))
	percentage = math.Round(percentage*multiplier) / multiplier

	// 格式化输出
	format := fmt.Sprintf("%%.%df%%%%", decimals)
	return fmt.Sprintf(format, percentage)
}
