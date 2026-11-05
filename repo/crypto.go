package repo

import (
	"context"
	"encoding/hex"
	"strconv"

	"go.uber.org/zap"

	"tongcheng/conf"
	"tongcheng/logs"
	"tongcheng/utils"
)

func EncryptUserInfo(ctx context.Context, userID int, plainText string) string {
	if len(plainText) == 0 {
		return ""
	}
	cipher, err := utils.AESEncrypt([]byte(plainText), utils.ReplaceAESKey(conf.GetConfiguration().Secret.InfoSecret, strconv.Itoa(userID)))
	if err != nil {
		logs.Error(ctx, "EncryptUserInfo AESEncrypt err", zap.Error(err))
	}
	return hex.EncodeToString(cipher)
}

func DecryptUserInfo(ctx context.Context, userID int, cipherText string) string {
	if len(cipherText) == 0 {
		return ""
	}
	cipher, err := hex.DecodeString(cipherText)
	if err != nil {
		logs.Error(ctx, "DecryptUserInfo DecodeString err", zap.Error(err))
	}
	plaintext, err := utils.AESDecrypt(cipher, utils.ReplaceAESKey(conf.GetConfiguration().Secret.InfoSecret, strconv.Itoa(userID)))
	if err != nil {
		logs.Error(ctx, "DecryptUserInfo AESDecrypt err", zap.Error(err))
	}
	return string(plaintext)
}
