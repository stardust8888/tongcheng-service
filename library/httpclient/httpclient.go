package httpclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"tongcheng/logs"
	"tongcheng/utils/constant"
	"tongcheng/utils/tool/jsoniter"
)

var Client *http.Client

func Do(ctx context.Context, module any, url, method string, headers map[string]string, body []byte) error {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		logs.Error(ctx, "HttpClient NewRequest err", zap.Error(err))
		return err
	}

	req.Header.Set(constant.HeaderContentType, constant.ContentTypeJSON)

	if headers == nil {
		headers = map[string]string{}
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	logs.Info(ctx, "HttpClient Do", zap.String("url", url), zap.String("body", string(body)))
	resp, err := Client.Do(req)
	if err != nil {
		logs.Error(ctx, "HttpClient Do err", zap.Error(err))
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logs.Error(ctx, "HttpClient Code err", zap.Int("code", resp.StatusCode))
		return errors.New(constant.GetCodeText(constant.CodeHTTPError))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.Error(ctx, "HttpClient ReadAll err", zap.Error(err))
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	err = jsoniter.Unmarshal(data, module)
	if err != nil {
		logs.Error(ctx, "HttpClient Unmarshal err", zap.Error(err))
		return err
	}

	return nil
}

func EmptyHeader() map[string]string {
	return map[string]string{}
}

func EmptyBody() []byte {
	return make([]byte, 0)
}

type HttpClient struct {
	client *http.Client
}

func CreateHttpClient() *HttpClient {
	transport := &http.Transport{
		MaxIdleConns:        90,
		MaxIdleConnsPerHost: 30,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		//TLSClientConfig:       config.TLSClientConfig,
		ExpectContinueTimeout: time.Second,
	}

	return &HttpClient{&http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}}
}

func (cli *HttpClient) Do(ctx context.Context, module any, url, method string, headers map[string]string, body []byte) error {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		logs.Error(ctx, "HttpClient NewRequest err", zap.Error(err))
		return err
	}

	req.Header.Set(constant.HeaderContentType, constant.ContentTypeJSON)

	if headers == nil {
		headers = map[string]string{}
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	logs.Info(ctx, "HttpClient Do", zap.String("url", url), zap.Int("bodySize", len(body)))
	resp, err := cli.client.Do(req)
	if err != nil {
		logs.Error(ctx, "HttpClient Do err", zap.Error(err))
		return err
	}

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		logs.Error(ctx, "HttpClient ReadAll err", zap.Error(err))
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		logs.Error(ctx, "HttpClient Code err", zap.Int("code", resp.StatusCode), zap.String("body", string(respBody)))
		return errors.New(constant.GetCodeText(constant.CodeHTTPError))
	}

	err = jsoniter.Unmarshal(respBody, module)
	if err != nil {
		logs.Error(ctx, "HttpClient Unmarshal err", zap.Error(err))
		return err
	}

	return nil
}
