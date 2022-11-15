package http_client

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/golang-infrastructure/go-cli-weixin-login-sdk/internal/logger"
	"time"
)

var RequestFailedError = errors.New("请求失败")

// GetAsString 获取给定页面的内容，以文本方式返回
func GetAsString(urlString string) (string, error) {
	for tryTimes := 1; tryTimes <= 3; tryTimes++ {
		resp, err := resty.
			New().
			SetTimeout(time.Second*30).
			R().
			SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36").
			SetHeader("Accept-Encoding", "gzip, deflate").
			SetHeader("Accept", "*/*").
			SetHeader("Accept-Language", "zh-CN,zh;q=0.9").
			SetHeader("Referer", "https://open.weixin.qq.com/").
			Get(urlString)

		if err != nil {
			logger.Error("请求失败，url = %s, error msg = %s, tryTimes = %d", urlString, err.Error(), tryTimes)
			continue
		}
		if resp.StatusCode() != 200 {
			logger.Error("请求失败，url = %s, 响应状态码不为200, tryTimes = %d", urlString, tryTimes)
			continue
		}
		return string(resp.Body()), nil
	}
	return "", RequestFailedError
}
