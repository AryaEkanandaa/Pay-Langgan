package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateBusinessID() string {
	return fmt.Sprintf("biz_%s", randomString(16))
}

func GenerateInvoiceNumber() string {
	now := time.Now()
	return fmt.Sprintf("INV-%s-%s", now.Format("20060102"), randomString(8))
}

func randomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
