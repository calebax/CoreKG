package svcforestfile

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestCalcOptimalPart(t *testing.T) {
	// 使用局部随机生成器
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	const numTests = 20000
	// 定义最小和最大值
	const (
		MinSize       int64 = 1 * 1024 * 1024           // 1MB
		MaxCommonSize int64 = 1 * 1024 * 1024 * 1024    // 1GB
		MaxSize       int64 = 500 * 1024 * 1024 * 10000 // 500MB * 10000 = 50GB
	)

	// Print table header
	fmt.Printf("%-5s %-14s %-14s %-10s\n", "No.", "FileSize", "PartSize", "PartCount")
	fmt.Printf("%s\n", "--------------------------------------------------------")

	for i := 0; i < numTests; i++ {
		// 随机生成文件大小，范围
		var fileSize int64
		if i == 0 {
			// 第一次测试最大文件大小
			fileSize = MaxSize
		} else if i > 10000 {
			// i > 10000 时使用常见大小，最大 1GB
			fileSize = r.Int63n(MaxCommonSize-MinSize+1) + MinSize
		} else {
			// 随机生成文件大小
			fileSize = r.Int63n(MaxSize-MinSize+1) + MinSize
		}

		partSize, partCount := CalcOptimalPart(fileSize)

		if partCount == 0 {
			t.Fatalf("Test #%d: partCount %d for fileSize %s", i+1, partCount, formatSize(fileSize))
		}
		if partCount > 1 {
			// 1️⃣ 检查 partSize 是否在合理范围
			if partSize < 5*1024*1024 {
				t.Fatalf("Test #%d: partSize %s < MinPartSize for fileSize %s", i+1, formatSize(partSize), formatSize(fileSize))
			}
			if partSize > 500*1024*1024 {
				t.Fatalf("Test #%d: partSize %s > MaxPartSize for fileSize %s", i+1, formatSize(partSize), formatSize(fileSize))
			}

			// 2️⃣ 检查 partCount 是否正确 (ceil 计算)
			if int64(partCount)*partSize < fileSize {
				t.Fatalf("Test #%d: partCount %d too small for fileSize %s with partSize %s", i+1, partCount, formatSize(fileSize), formatSize(partSize))
			}

			// 3️⃣ 检查是否超出最大分片数
			if partCount > 10000 {
				t.Fatalf("Test #%d: partCount %d exceeds MaxParts for fileSize %s", i+1, partCount, formatSize(fileSize))
			}
		}

		fmt.Printf("%-5d %-14s %-14s %-10d\n",
			i,
			formatSize(fileSize),
			formatSize(partSize),
			partCount,
		)
	}
}

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	default:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	}
}
