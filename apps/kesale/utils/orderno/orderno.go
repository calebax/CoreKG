package orderno

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	mu                   sync.Mutex
	currentMinute        string
	usedNumbers          map[int]struct{}
	currentPaymentMinute string
	usedPaymentNumbers   map[int]struct{}
)

func Generate(bizKey string, machineID int) string {
	mid := fmt.Sprintf("%01d", machineID%10)
	minute := time.Now().UTC().Format("0601021504")

	var randNum int
	for {
		randNum = rand.Intn(1000000) // 0~999999

		mu.Lock()
		if currentMinute != minute {
			currentMinute = minute
			usedNumbers = make(map[int]struct{})
		}

		if _, exists := usedNumbers[randNum]; !exists {
			usedNumbers[randNum] = struct{}{}
			mu.Unlock()
			break
		}
		mu.Unlock()
	}

	r := fmt.Sprintf("%06d", randNum)

	p := sanitize(bizKey)
	if len(p) == 0 {
		p = "KG"
	}

	return p + minute + mid + r
}

func sanitize(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
		if b.Len() >= 8 {
			break
		}
	}
	return strings.ToUpper(b.String())
}

// GeneratePaymentTradeNo 生成20位支付交易号
func GeneratePaymentTradeNo() string {
	timestamp := time.Now().UTC().Format("060102150405")

	var randNum int
	for {
		randNum = rand.Intn(100000000) // 0~99999999

		mu.Lock()
		if currentPaymentMinute != timestamp {
			currentPaymentMinute = timestamp
			usedPaymentNumbers = make(map[int]struct{})
		}

		if _, exists := usedPaymentNumbers[randNum]; !exists {
			usedPaymentNumbers[randNum] = struct{}{}
			mu.Unlock()
			break
		}
		mu.Unlock()
	}

	r := fmt.Sprintf("%08d", randNum)
	return timestamp + r
}
