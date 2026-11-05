package googleauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

// GenerateSecret 生成一个用于Google Authenticator的密钥
// 密钥长度通常为16-32个字符，使用Base32编码
func GenerateSecret(length int) (string, error) {
	// 生成随机密钥
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	secret := make([]byte, length)

	// 使用crypto/rand生成真正的随机数
	for i := range secret {
		// 生成0-31的随机数
		b := make([]byte, 1)
		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}
		// 将随机字节映射到charset索引
		secret[i] = charset[int(b[0])%len(charset)]
	}

	return string(secret), nil
}

// GenerateCode 根据密钥和时间生成TOTP验证码
func GenerateCode(secret string, t time.Time) (string, error) {
	// 解码Base32密钥
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", err
	}

	// 计算时间计数器 (默认30秒一个周期)
	counter := t.Unix() / 30

	// 将计数器转换为字节数组
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	// 使用HMAC-SHA1算法计算哈希值
	h := hmac.New(sha1.New, key)
	h.Write(buf)
	hash := h.Sum(nil)

	// 动态截断
	offset := hash[len(hash)-1] & 0x0F
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF

	// 取模得到6位数字验证码
	code := truncated % 1000000

	// 格式化为6位数字字符串
	return fmt.Sprintf("%06d", code), nil
}

// ValidateCode 验证用户输入的TOTP代码
// 允许前后一个时间窗口的误差（默认±30秒）
func ValidateCode(secret, code string) (bool, error) {
	// 当前时间验证
	currentCode, err := GenerateCode(secret, time.Now())
	if err != nil {
		return false, err
	}

	if currentCode == code {
		return true, nil
	}

	// 验证前一个时间窗口
	prevCode, err := GenerateCode(secret, time.Now().Add(-30*time.Second))
	if err != nil {
		return false, err
	}

	if prevCode == code {
		return true, nil
	}

	// 验证后一个时间窗口
	nextCode, err := GenerateCode(secret, time.Now().Add(30*time.Second))
	if err != nil {
		return false, err
	}

	if nextCode == code {
		return true, nil
	}

	return false, nil
}

// GetQRCodeContent 生成用于二维码的内容（otpauth URL）
func GetQRCodeContent(secret, account, issuer string) string {
	// 返回otpauth URL格式，用于生成二维码
	return fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s", account, secret, issuer)
}

// IsValidBase32 检查密钥是否为有效的base32编码
func IsValidBase32(secret string) bool {
	_, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	return err == nil
}

// GetValidUntil 返回当前验证码的有效期结束时间
func GetValidUntil() time.Time {
	now := time.Now()
	// 默认30秒周期，计算当前周期的结束时间
	period := int64(30)
	currentPeriod := now.Unix() / period
	return time.Unix((currentPeriod+1)*period, 0)
}

// GetRemainingTime 返回当前验证码的剩余有效时间（秒）
func GetRemainingTime() int {
	now := time.Now()
	period := int64(30)
	currentPeriod := now.Unix() / period
	nextPeriod := (currentPeriod + 1) * period
	return int(nextPeriod - now.Unix())
}

// CalculateEntropy 计算密钥的熵值（用于评估密钥强度）
func CalculateEntropy(secret string) float64 {
	// Base32字符集大小为32
	charsetSize := 32.0
	// 熵值计算公式: log2(charsetSize^length) = length * log2(charsetSize)
	return float64(len(secret)) * math.Log2(charsetSize)
}
