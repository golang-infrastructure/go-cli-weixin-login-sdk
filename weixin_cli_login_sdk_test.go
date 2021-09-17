package go_cli_weixin_login_sdk

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestLogin(t *testing.T) {

	// https://open.weixin.qq.com/connect/qrcode/
	login, err := Login("wxb3df453f8a216de2")
	if err != nil {
		println(err.Error())
		return
	}
	marshal, _ := json.Marshal(login)
	fmt.Println(string(marshal))

}
