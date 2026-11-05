package repo

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"tongcheng/library/httpclient/wxapi"
	"tongcheng/logs"
	"tongcheng/model"
)

// checkChildText 对子女资料自由文本做微信内容安全检测(msg_sec_check)。
// description/interest 为 UGC 自由文本必检；company/work 为枚举值字符串附带检测。
// 命中违规返回 true(应拦截)；网络/API 错误记日志后放行(fail-open)，避免微信抖动阻断全量用户。
func CheckChildText(ctx context.Context, req *model.UserChild) bool {
	var parts []string
	for _, t := range []string{req.Description, req.Interest} {
		if strings.TrimSpace(t) != "" {
			parts = append(parts, t)
		}
	}
	if len(parts) == 0 {
		return false
	}
	content := strings.Join(parts, " ")
	risky, err := wxapi.GetClient().MsgSecCheck(ctx, content)
	if err != nil {
		logs.Error(ctx, "child text sec check fail-open", zap.Error(err))
		return false
	}
	if risky {
		logs.Error(ctx, "child text sec check risky", zap.String("content", content))
		return true
	}
	return false
}

// CheckMomentText 对动态正文做微信内容安全检测(msg_sec_check)。
// content 为 UGC 自由文本必检；空串直接放行。
// 命中违规返回 true(应拦截)；网络/API 错误记日志后放行(fail-open)，避免微信抖动阻断全量用户。
func CheckMomentText(ctx context.Context, content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}
	risky, err := wxapi.GetClient().MsgSecCheck(ctx, content)
	if err != nil {
		logs.Error(ctx, "moment text sec check fail-open", zap.Error(err))
		return false
	}
	if risky {
		logs.Error(ctx, "moment text sec check risky", zap.String("content", content))
		return true
	}
	return false
}
