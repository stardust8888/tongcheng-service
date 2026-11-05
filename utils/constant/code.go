package constant

const (
	// 通用错误码 10000-19999
	CodeSuccess        = 200   // 成功
	CodeUnknown        = 10000 // 未知错误
	CodeNotify         = 10001 // 前端提示
	CodeParamError     = 10002 // 参数错误
	CodeServerInternal = 10003 // 服务器内部错误
	CodeUnauthorized   = 10004 // 未授权
	CodeNotFound       = 10005 // 资源不存在
	CodeTimeout        = 10006 // 请求超时

	// 用户相关错误码 20000-29999
	CodeUserNotExist           = 20000
	CodeUserNotExistGoRegister = 20001
	CodeUserNoChild            = 20002
	CodeUserOneChildOnly       = 20003
	CodeUserDisabled           = 20004
	CodePasswordError          = 20005
	CodePermissionDenied       = 20013 // 权限不足

	CodeDataAlreadyExists = 30002 // 数据已存在
	CodeInvalidOperation  = 30009 // 无效操作

	CodeHTTPError = 60003 // HTTP错误

	// 文件相关错误码 80000-89999
	CodeFileError         = 80000 // 文件错误
	CodeFileNotFound      = 80001 // 文件未找到
	CodeFileUploadError   = 80002 // 文件上传错误
	CodeFileDownloadError = 80003 // 文件下载错误
	CodeFileSizeExceeded  = 80004 // 文件大小超限
	CodeFileFormatInvalid = 80005 // 文件格式无效
	CodeFilePermission    = 80006 // 文件权限错误
	CodeFileCannotRead    = 80007 // 文件无法读取

	// 消息推送相关错误码 100000-109999
	CodeMessageTypeError       = 100000
	CodeMessageContentError    = 100001
	CodeMessageSendError       = 100002
	CodeMessageReceiveError    = 100003
	CodeMessageTimeoutError    = 100004
	CodeMessageCreateChatError = 100005
)

var CodeMap = map[int]string{
	CodeSuccess:        "成功",
	CodeUnknown:        "未知",
	CodeParamError:     "参数错误",
	CodeUnauthorized:   "未授权",
	CodeNotFound:       "资源不存在或已被删除",
	CodeTimeout:        "请求超时",
	CodeServerInternal: "服务器繁忙",

	// 用户相关错误
	CodeUserNotExist:           "您未注册账户",
	CodeUserNotExistGoRegister: "请您注册账户",
	CodeUserNoChild:            "您未添加子女",
	CodeUserOneChildOnly:       "抱歉，您只能添加一位子女",
	CodeUserDisabled:           "用户被禁用",
	CodePasswordError:          "密码错误",
	CodePermissionDenied:       "权限不足",
	CodeDataAlreadyExists:      "数据已存在",
	CodeInvalidOperation:       "无效操作",

	CodeHTTPError: "HTTP错误",

	// 文件相关错误
	CodeFileError:         "文件错误",
	CodeFileNotFound:      "文件未找到",
	CodeFileUploadError:   "文件上传错误",
	CodeFileDownloadError: "文件下载错误",
	CodeFileSizeExceeded:  "文件大小超限",
	CodeFileFormatInvalid: "文件格式无效",
	CodeFilePermission:    "文件权限错误",
	CodeFileCannotRead:    "文件无法读取",

	// 消息推送相关错误
	CodeMessageTypeError:       "消息类型错误",
	CodeMessageContentError:    "消息内容错误",
	CodeMessageSendError:       "消息发送错误",
	CodeMessageReceiveError:    "消息接收错误",
	CodeMessageTimeoutError:    "消息超时",
	CodeMessageCreateChatError: "创建聊天错误",
}

func GetCodeText(code int) string {
	if text, ok := CodeMap[code]; ok {
		return text
	}
	return CodeMap[CodeUnknown]
}
