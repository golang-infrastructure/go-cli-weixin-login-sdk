package http_client

import (
	"errors"
	"github.com/CC11001100/go-cli-weixin-login-sdk/internal/logger"
	"github.com/go-resty/resty/v2"
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
			SetHeader("user-agent", "Mozilla / 5.0 (Windows NT 10.0; WOW64) AppleWebKit / 537.36 (KHTML, like Gecko) Chrome / 63.0.3239.132 Safari / 537.36 (WenZaiZhiBoClient-Windows7-weishi-8.6.6)").
			SetHeader("accept-encoding", "gzip, deflate").
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
