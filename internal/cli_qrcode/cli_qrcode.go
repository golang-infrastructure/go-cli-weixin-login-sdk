package cli_qrcode

import (
	"bytes"
	"github.com/CC11001100/go-cli-weixin-login-sdk/internal/logger"
	"github.com/skip2/go-qrcode"
	"image/png"
	"strings"
)

// ShowQRCodeOnCLI 在命令行展示二维码，供扫码登录用
func ShowQRCodeOnCLI(content string) {
	size := 41

	var code []byte
	code, err := qrcode.Encode(content, qrcode.Low, size)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	buf := bytes.NewBuffer(code)
	img, err := png.Decode(buf)
	if err != nil {
		return
	}

	// 先组装然后再打印，避免被弄乱
	sb := strings.Builder{}
	sb.WriteString("\n")
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			r32, g32, b32, _ := img.At(x, y).RGBA()
			r, g, b := int(r32>>8), int(g32>>8), int(b32>>8)
			if (r+g+b)/3 > 180 {
				sb.WriteString("##")
			} else {
				sb.WriteString("  ")
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	logger.White(sb.String())
}
