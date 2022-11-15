package go_cli_weixin_login_sdk

import (
	"errors"
	"fmt"
	"github.com/golang-infrastructure/go-cli-weixin-login-sdk/internal/cli_qrcode"
	"github.com/golang-infrastructure/go-cli-weixin-login-sdk/internal/http_client"
	"github.com/golang-infrastructure/go-cli-weixin-login-sdk/internal/logger"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LoginResult 扫码登录的结果
type LoginResult struct {

	// 大多数情况下只需要关注这两个字段就可以了，先检查是否登录成功，如果成功的话就使用code
	// 是否登录成功
	IsLoginSuccess bool
	// 登录成功后的code
	Code string

	// 下面是一些零碎的状态，只是为了让外界能够获取得到
	// 登录时使用的UUID
	UUID string
	// 状态码
	StatusCode int
	// 对状态码的解释
	StatusCodeMsg string
}

// checkResult 检查登录二维码被扫的状态
type checkResult struct {

	// 表示二维码的当前状态，比如是不是过期了，是不是被扫描过了
	StatusCode int

	// 登录成功才会有，这个是登录成功之后发给的凭据
	SuccessCode string
}

// 在长轮询检查微信登录的时候，响应的每种状态码的含义
var wxErrCodeNameMap = make(map[int]string, 0)

func init() {
	wxErrCodeNameMap[402] = "二维码过期"
	wxErrCodeNameMap[500] = wxErrCodeNameMap[402]
	wxErrCodeNameMap[403] = "用户取消登录"
	wxErrCodeNameMap[404] = "用户扫描二维码成功"
	wxErrCodeNameMap[405] = "登录成功"
	wxErrCodeNameMap[408] = "等待登陆"
}

// Login 尝试登录某个app
func Login(appId string) (*LoginResult, error) {

	// 1. 获取login uuid
	logger.Info("开始获取微信登录UUID...")
	loginUUID, err := getAppLoginUUID(appId)
	if err != nil {
		logger.Error("获取微信登录UUID失败，error msg = %s", err.Error())
		return nil, err
	}
	logger.Info("获取微信登录UUID成功，UUID = %s", loginUUID)

	// 2. 展示登录二维码
	loginUrl := "https://open.weixin.qq.com/connect/confirm?uuid=" + loginUUID
	cli_qrcode.ShowQRCodeOnCLI(loginUrl)

	// 注意，从这里开始要自己换行了
	logger.White("请微信扫码登录...")

	// 3. 轮询等待，直到扫描成功或者二维码过期
	isStopCheck := false
	var checkResult *checkResult
	scanDoneShowMessage := "\n掏出手机，扫描了二维码，请授权微信登录..."
	errorCount := 0
	for !isStopCheck {
		urlString := fmt.Sprintf("https://lp.open.weixin.qq.com/connect/l/qrconnect?uuid=%s&_=%d", loginUUID, time.Now().Unix())
		responseBody, err := http_client.GetAsString(urlString)
		if err != nil {
			// 累积错误退出防卡死
			errorCount++
			if errorCount >= 10 {
				isStopCheck = true
			} else {
				time.Sleep(time.Second * 1)
			}
			continue
		}
		// 等待登录
		// window.wx_errcode=408;window.wx_code='';
		// 登录成功
		// window.wx_errcode=405;window.wx_code='031MdHkl2MwgL74MxSol23J3fk2MdHkK';
		// 不同errcode的含义：
		// 500 过期
		// 402 过期
		// 403 取消登录
		// 404 扫描成功
		// 405 登录成功
		// 408 等待登陆
		checkResult = ParseWxLoginCheckResponse(responseBody)
		if checkResult == nil {
			isStopCheck = true
			continue
		}
		switch checkResult.StatusCode {
		case 402:
		case 500:
			// 二维码过期
			logger.Red("\n二维码已过期，登录失败！")
			isStopCheck = true
			break
		case 403:
			// 取消登录
			logger.Red("\n拒绝授权，登录失败！")
			isStopCheck = true
			break
		case 404:
			// 扫描成功
			logger.Green(scanDoneShowMessage)
			scanDoneShowMessage = "."
			time.Sleep(time.Millisecond * 500)
			break
		case 405:
			// 登录成功
			logger.Green("\n微信授权成功！")
			isStopCheck = true
			break
		case 408:
			// 等待登录，就打印个.表示还在检查
			logger.White("...")
			break
		}
	}
	logger.Info("")

	// 组装结果返回
	loginResult := LoginResult{
		UUID: loginUUID,
	}

	if checkResult != nil {
		loginResult.Code = checkResult.SuccessCode
		loginResult.StatusCode = checkResult.StatusCode
		loginResult.StatusCodeMsg = wxErrCodeNameMap[checkResult.StatusCode]
	}

	if loginResult.Code != "" {
		loginResult.IsLoginSuccess = true
	}

	return &loginResult, nil
}

func ParseWxLoginCheckResponse(responseBody string) *checkResult {
	split := strings.Split(responseBody, ";")
	if len(split) != 3 {
		return nil
	}
	statusSplit := strings.Split(split[0], "=")
	if len(statusSplit) != 2 {
		return nil
	}
	status, _ := strconv.Atoi(statusSplit[1])
	codeSplit := strings.Split(split[1], "=")
	if len(codeSplit) != 2 {
		return nil
	}
	return &checkResult{
		StatusCode:  status,
		SuccessCode: strings.Trim(codeSplit[1], "'"),
	}
}

var GetLoginUUIDFailedError = errors.New("获取login UUID失败")

// 获取某个应用的登录id
func getAppLoginUUID(appId string) (string, error) {
	urlString := "https://open.weixin.qq.com/connect/qrconnect?appid=" + appId + "&scope=snsapi_login&redirect_uri=https://&state=STATE&login_type=jssdk&self_redirect=default&styletype=&sizetype=&bgcolor=&rst=&style=black&href="
	responseBody, err := http_client.GetAsString(urlString)
	if err != nil {
		return "", err
	}

	compile, err := regexp.Compile("\"https://long.open.weixin.qq.com/connect/l/qrconnect\\?uuid=(.+?)\"")
	if err != nil {
		return "", GetLoginUUIDFailedError
	}
	match := compile.FindStringSubmatch(responseBody)
	if len(match) < 2 {
		return "", GetLoginUUIDFailedError
	}
	return match[1], nil
}
