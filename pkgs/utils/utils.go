package utils

import (
	"time"
)

// Retry 重試
func Retry(attempts int, sleep int, fn func() (error, bool)) (err error) {
	var needContinue bool
	for i := 0; i < attempts; i++ {
		err, needContinue = fn()
		if !needContinue {
			return err
		}
		if sleep > 0 {
			time.Sleep(time.Duration(sleep) * time.Second)
		}
	}
	return
}

func GetUnixMicro(duration time.Duration) int64 {
	return time.Now().Add(duration).UnixMicro()
}

// SliceDuplicate 切片去重
func SliceDuplicate[T string | uint](list []T) []T {
	m := make(map[T]struct{})
	result := make([]T, 0, len(list)) // 创建一个新的切片来存储结果
	for _, v := range list {
		if _, ok := m[v]; !ok {
			m[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// ToMap
// K: Map 的键类型 (Key type)。它必须是 Go 中可作为 map 键的类型，
// 确保 K 是可比较的类型，这是 map 键的要求。
//
// 使用 comparable 约束。
//
// V: 切片和 Map 中结构体的值类型 (Value type)。它必须是一个指针类型（any 是 Go 1.18+ 的别名）
func ToMap[K comparable, V any](
	slice []V,
	keyFunc func(V) K,
) map[K]V {
	// 1. 初始化 Map
	// 预分配容量，优化性能。如果 slice 的长度已知，分配相同长度的 map 可以减少后续的 rehash。
	resultMap := make(map[K]V, len(slice))

	// 2. 遍历切片并填充 Map
	for _, val := range slice {
		// 3. 使用 keyFunc 获取键
		// 调用用户传入的 keyFunc 函数，从当前的结构体 V 中提取出 Map 的键 K。
		key := keyFunc(val)

		// 4. 赋值
		// 将当前元素（V）以提取出的键（K）存入 Map。
		resultMap[key] = val
	}

	return resultMap
}

// PtrValue 安全获取指针指向的值
// 如果指针为 nil，返回该类型的零值 (例如 string 返回 "", int 返回 0, bool 返回 false)
// 如果指针不为 nil，返回指针指向的实际值
func PtrValue[T any](p *T) T {
	if p == nil {
		var zero T // 声明一个该类型的变量，默认就是零值
		return zero
	}
	return *p
}

// Map 通用转换函数
// slice: 要转换的切片
// iteratee: 转换函数
// 返回转换后的切片
func Map[T any, R any](slice []T, iteratee func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = iteratee(v)
	}
	return result
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
