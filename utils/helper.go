package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"tongcheng/utils/tool/jsoniter"
	"tongcheng/utils/tool/spew"
)

func ToRawMessage(i any) json.RawMessage {
	b := ToJsonByte(i)
	return b
}

func ToJsonByte(i any) []byte {
	b, _ := jsoniter.Marshal(i)
	return b
}

func ToJson(i any) string {
	b, _ := jsoniter.Marshal(i)
	return string(b)
}

func ToString(i any) string {
	return fmt.Sprint(i)
}

func Spew(i any) {
	spew.Dump(i)
}

// RandString 生成指定长度的随机字符串
func RandString(length int) string {
	// 字符集，可以根据需要修改
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func RandNumberString(length int) string {
	// 字符集，可以根据需要修改
	charset := "012356789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// RandStringWithCharset 使用指定字符集生成指定长度的随机字符串
func RandStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
