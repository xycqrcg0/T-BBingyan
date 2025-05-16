package util

import (
	"BBingyan/internal/config"
	"fmt"
	"github.com/jordan-wright/email"
	"math/rand"
	"net/smtp"
)

func GenerateCode() string {
	return fmt.Sprintf("%4d", rand.Int()%9999)
}

func SendAuthCode(to string, code string) error {
	e := email.NewEmail()
	e.From = "BBingyan <bbingyan@qq.com>"
	e.To = []string{to}
	e.Subject = "Auth Code"
	text := fmt.Sprintf("【BBingyan】验证码: %s ,\n请在5分钟内进行验证,如非本人操作请忽略.\n感谢使用BBingyan!", code)
	e.Text = []byte(text)

	return e.Send("smtp.qq.com:25", smtp.PlainAuth("", "bbingyan@qq.com", config.Config.AuthorizationCode, "smtp.qq.com"))
}
