package constant

const (
	BaseWebUrlV1   = "/web/v1"
	BaseAdminUrlV1 = "/admin/v1"
	BaseWsUrlV1    = "/ws/v1"
	BaseFileUrlV1  = "/file/v1"
	BaseCronUrlV1  = "/cron/v1"
)

type SimpleUser struct {
	ID       int
	WxOpenID string
	IsAdmin  int
}

const (
	SuperAdmin = "超级管理员"
	UserAdmin  = "用户管理员"

	PrefixUsername = "家长"
)

const (
	UserID    = "UserID"
	WxOpenID  = "WxOpenID"
	AdminID   = "AdminID"
	HeartBeat = "HeartBeat"
)

const (
	MaskedString = "******"
)

const (
	B  = 1
	KB = 1024
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
	TB = 1024 * 1024 * 1024 * 1024
)

// HTTP常用Content-Type
const (
	ContentTypeJSON = "application/json"
)

// HTTP常用请求头键名
const (
	HeaderAuthTokenInURL = "token"
	HeaderAuthorization  = "Authorization"
	HeaderContentType    = "Content-Type"
	HeaderContentLength  = "Content-Length"
	HeaderAccept         = "Accept"
)
