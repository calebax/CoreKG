package types

import "fmt"

type ByteSize int64

func (b ByteSize) String() string {
	if b < 1024 {
		return fmt.Sprintf("%dB", b)
	}
	if b < 1024*1024 {
		return fmt.Sprintf("%.2fK", float64(b)/1024)
	}
	if b < 1024*1024*1024 {
		return fmt.Sprintf("%.2fM", float64(b)/(1024*1024))
	}
	if b < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2fG", float64(b)/(1024*1024*1024))
	}
	return fmt.Sprintf("%.2fT", float64(b)/(1024*1024*1024*1024))
}
