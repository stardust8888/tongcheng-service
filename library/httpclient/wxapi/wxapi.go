package wxapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"tongcheng/conf"
	"tongcheng/library/httpclient"
	"tongcheng/library/snowflake"
	"tongcheng/logs"
	"tongcheng/utils"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

var (
	client *Client
)

type Client struct {
	AccessToken string
	ExpiresAt   int
	Config      conf.WxConfig
	HttpClient  *httpclient.HttpClient
}

func InitClient() {
	client = &Client{
		Config:     conf.GetConfiguration().Wx,
		HttpClient: httpclient.CreateHttpClient(),
	}
}

func GetClient() *Client {
	return client
}

func (wx *Client) IsExpired() bool {
	return wx.ExpiresAt < int(timex.GetBeijingTime().Unix())
}

/*
1.获取token
2.快速登录
3.发送短信
4.获取手机号
5.获取用户信息
6.消息通知
7.微信下单  /v3/pay/transactions/jsapi --> prepay_id
小程序调起支付 wx.requestPayment
8.订单查询 /v3/pay/transactions/out-trade-no/{out_trade_no}
9.回调接口  /web/v1/wxpay/callback
*/
type WxCommon struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type AccessTokenResp struct {
	WxCommon
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (wx *Client) GetAccessToken(ctx context.Context) error {
	var (
		resp  AccessTokenResp
		wxUrl = wx.Config.Endpoint + "/cgi-bin/token?"
		err   error
	)

	params := url.Values{}
	params.Add("appid", wx.Config.AppID)
	params.Add("secret", wx.Config.Secret)
	params.Add("grant_type", "client_credential")
	wxUrl += params.Encode()
	err = wx.HttpClient.Do(ctx, &resp, wxUrl, http.MethodGet, httpclient.EmptyHeader(), httpclient.EmptyBody())
	if err != nil {
		return err
	}

	if resp.ErrCode != 0 {
		logs.Error(ctx, "WxAccessToken err", zap.Int("code", resp.ErrCode), zap.String("msg", resp.ErrMsg))
		return errors.New(resp.ErrMsg)
	}

	wx.AccessToken = resp.AccessToken
	wx.ExpiresAt = int(timex.GetBeijingTime().Unix()) + resp.ExpiresIn - 60
	return nil
}

type GetOpenIDResp struct {
	WxCommon
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
}

func (wx *Client) GetOpenID(ctx context.Context, code string) (*GetOpenIDResp, error) {
	var (
		resp  GetOpenIDResp
		wxUrl = wx.Config.Endpoint + "/sns/jscode2session?"
		err   error
	)

	if code == "" {
		return nil, errors.New(constant.GetCodeText(constant.CodeParamError))
	}

	if wx.IsExpired() {
		err = wx.GetAccessToken(ctx)
		if err != nil {
			return nil, err
		}
	}

	params := url.Values{}
	params.Add("appid", wx.Config.AppID)
	params.Add("secret", wx.Config.Secret)
	params.Add("js_code", code)
	params.Add("grant_type", "authorization_code")
	wxUrl += params.Encode()
	err = wx.HttpClient.Do(ctx, &resp, wxUrl, http.MethodGet, httpclient.EmptyHeader(), httpclient.EmptyBody())
	if err != nil {
		return nil, err
	}

	if resp.ErrCode != 0 {
		logs.Error(ctx, "WxFastLogin err", zap.Int("code", resp.ErrCode), zap.String("msg", resp.ErrMsg))
		return nil, errors.New(resp.ErrMsg)
	}

	return &resp, nil
}

type GetUserPhoneNumberResp struct {
	WxCommon
	PhoneInfo PhoneInfo `json:"phone_info"`
}
type PhoneInfo struct {
	PhoneNumber     string    `json:"phoneNumber"`
	PurePhoneNumber string    `json:"purePhoneNumber"`
	CountryCode     string    `json:"countryCode"`
	Watermark       Watermark `json:"watermark"`
}
type Watermark struct {
	AppID     string `json:"appid"`
	Timestamp int    `json:"timestamp"`
}

func (wx *Client) GetUserPhoneNumber(ctx context.Context, code string) (*GetUserPhoneNumberResp, error) {
	var (
		resp  GetUserPhoneNumberResp
		wxUrl = wx.Config.Endpoint + "/wxa/business/getuserphonenumber?"
		err   error
	)

	if wx.IsExpired() {
		err = wx.GetAccessToken(ctx)
		if err != nil {
			return nil, err
		}
	}

	params := url.Values{}
	params.Add("access_token", wx.AccessToken)
	wxUrl += params.Encode()
	body := map[string]any{
		"code": code,
	}
	err = wx.HttpClient.Do(ctx, &resp, wxUrl, http.MethodPost, httpclient.EmptyHeader(), utils.ToJsonByte(body))
	if err != nil {
		return nil, err
	}

	if resp.ErrCode != 0 {
		logs.Error(ctx, "GetUserPhoneNumber err", zap.Int("code", resp.ErrCode), zap.String("msg", resp.ErrMsg))
		return nil, errors.New(resp.ErrMsg)
	}

	return &resp, nil
}

type PrepayRequest struct {
	AppID         string     `json:"appid"`        //必填
	MchID         string     `json:"mchid"`        //必填
	Description   string     `json:"description"`  //必填
	OutTradeNo    string     `json:"out_trade_no"` //必填
	TimeExpire    timex.Time `json:"time_expire"`  //必填
	Attach        string     `json:"attach"`
	NotifyURL     string     `json:"notify_url"` //必填
	GoodsTag      string     `json:"goods_tag"`
	SupportFapiao bool       `json:"support_fapiao"`
	Amount        Amount     `json:"amount"` //必填
	Payer         Payer      `json:"payer"`  //必填
	Detail        Detail     `json:"detail"`
	SceneInfo     SceneInfo  `json:"scene_info"`
	SettleInfo    SettleInfo `json:"settle_info"`
}

type Amount struct {
	Total    float64 `json:"total"`
	Currency string  `json:"currency"`
}

type Payer struct {
	Openid string `json:"openid"`
}

type GoodsDetail struct {
	MerchantGoodsID  string `json:"merchant_goods_id"`
	WechatpayGoodsID string `json:"wechatpay_goods_id"`
	GoodsName        string `json:"goods_name"`
	Quantity         int    `json:"quantity"`
	UnitPrice        int    `json:"unit_price"`
}

type Detail struct {
	CostPrice   int           `json:"cost_price"`
	InvoiceID   string        `json:"invoice_id"`
	GoodsDetail []GoodsDetail `json:"goods_detail"`
}

type StoreInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	AreaCode string `json:"area_code"`
	Address  string `json:"address"`
}

type SceneInfo struct {
	PayerClientIP string    `json:"payer_client_ip"`
	DeviceID      string    `json:"device_id"`
	StoreInfo     StoreInfo `json:"store_info"`
}

type SettleInfo struct {
	ProfitSharing bool `json:"profit_sharing"`
}

type PrepayResponse struct {
	PrepayID string `json:"prepay_id"`
}

func (wx *Client) PrePay(ctx context.Context, req *PrepayRequest) (*PrepayResponse, error) {
	var (
		resp  PrepayResponse
		wxUrl = wx.Config.PayEndpoint + "/v3/pay/transactions/jsapi?"
		err   error
	)

	req.AppID = wx.Config.AppID
	req.MchID = wx.Config.MchID
	req.OutTradeNo = fmt.Sprintf("wx%v", snowflake.Generate())
	req.TimeExpire = timex.GetBeijingTime().Add(15 * timex.Minute)
	req.NotifyURL = conf.GetConfiguration().Wx.PayEndpointCallback + "/web/v1/pay/wx/callback"
	header := map[string]string{
		constant.HeaderAccept:        constant.ContentTypeJSON,
		constant.HeaderAuthorization: "",
	}
	err = wx.HttpClient.Do(ctx, &resp, wxUrl, http.MethodPost, header, utils.ToJsonByte(req))
	if err != nil {
		return nil, err
	}
	if resp.PrepayID == "" {
		logs.Error(ctx, "WxPrePay err", zap.String("response", utils.ToJson(resp)))
		return nil, errors.New("微信预支付失败")
	}

	return &resp, nil
}

// MsgSecCheck 文本内容安全检测（security.msg_sec_check）。
// content 为待检文本，空串直接放行。
// 返回 (risky, err)：risky=true 表示命中违规(errcode 87014)应拦截；
// err!=nil 为网络或微信 API 层错误，调用方应记日志后放行(fail-open)，避免微信抖动阻断全量用户。
func (wx *Client) MsgSecCheck(ctx context.Context, content string) (bool, error) {
	if content == "" {
		return false, nil
	}
	if wx.IsExpired() {
		if err := wx.GetAccessToken(ctx); err != nil {
			return false, err
		}
	}
	wxUrl := wx.Config.Endpoint + "/wxa/msg_sec_check?"
	params := url.Values{}
	params.Add("access_token", wx.AccessToken)
	wxUrl += params.Encode()
	body := map[string]any{"content": content}
	var resp WxCommon
	err := wx.HttpClient.Do(ctx, &resp, wxUrl, http.MethodPost, httpclient.EmptyHeader(), utils.ToJsonByte(body))
	if err != nil {
		return false, err
	}
	if resp.ErrCode == 87014 {
		logs.Error(ctx, "WxMsgSecCheck risky", zap.Int("code", resp.ErrCode), zap.String("msg", resp.ErrMsg))
		return true, nil
	}
	if resp.ErrCode != 0 {
		logs.Error(ctx, "WxMsgSecCheck api err", zap.Int("code", resp.ErrCode), zap.String("msg", resp.ErrMsg))
		return false, errors.New(resp.ErrMsg)
	}
	return false, nil
}

// ImgSecCheck 图片内容安全检测（security.img_sec_check）。
// media 为图片字节，filename 用于 WeChat 识别 media 字段文件名(带扩展名)。
// 返回 (risky, err)，语义同 MsgSecCheck：risky=true 拦截，err!=nil 放行。
// 走 HttpClient.DoSilent(绑定 ctx、不记二进制 body)，复用封装 Client 的超时与取消能力。
func (wx *Client) ImgSecCheck(ctx context.Context, media []byte, filename string) (bool, error) {
	if len(media) == 0 {
		return false, nil
	}
	// img_sec_check 图片上限约 1MB；超限前端已压缩故极少，超限直接放行并记录
	const imgMaxSize = 1 << 20
	if len(media) > imgMaxSize {
		logs.Error(ctx, "WxImgSecCheck skip oversized", zap.Int("size", len(media)))
		return false, nil
	}
	if wx.IsExpired() {
		if err := wx.GetAccessToken(ctx); err != nil {
			return false, err
		}
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("media", filename)
	if err != nil {
		return false, err
	}
	if _, err = fw.Write(media); err != nil {
		return false, err
	}
	if err = mw.Close(); err != nil {
		return false, err
	}
	wxUrl := wx.Config.Endpoint + "/wxa/img_sec_check?access_token=" + url.QueryEscape(wx.AccessToken)
	// 自定义 Content-Type 覆盖 Do 默认的 JSON，改为 multipart/form-data; boundary=...
	header := map[string]string{
		constant.HeaderContentType: mw.FormDataContentType(),
	}
	var sec WxCommon
	if err = wx.HttpClient.Do(ctx, &sec, wxUrl, http.MethodPost, header, buf.Bytes()); err != nil {
		return false, err
	}
	if sec.ErrCode == 87014 {
		logs.Error(ctx, "WxImgSecCheck risky", zap.Int("code", sec.ErrCode), zap.String("msg", sec.ErrMsg))
		return true, nil
	}
	if sec.ErrCode != 0 {
		logs.Error(ctx, "WxImgSecCheck api err", zap.Int("code", sec.ErrCode), zap.String("msg", sec.ErrMsg))
		return false, errors.New(sec.ErrMsg)
	}
	return false, nil
}
