package email

import (
	"github.com/adriamb/bcdapp/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/smtp"
	"net/url"	
	emailer "github.com/jordan-wright/email"
)

func AuthCode(address, email string) string {
	mac := hmac.New(sha256.New, []byte(config.C.ServerSecret))
	mac.Write([]byte(address))
	mac.Write([]byte(email))
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum)
}

func SendAuthEmail(address, email string) error {

	auth := smtp.PlainAuth(
		"",
		config.C.SmtpClient.User,
		config.C.SmtpClient.Password,
		config.C.SmtpClient.Domain,
	)

	e := emailer.NewEmail()
	e.Headers.Add("Content-Transfer-Encoding","quoted-printable")
	e.From = config.C.SmtpClient.From
	e.To = []string{email}
	e.Subject = "Blockchain Catalunya - Verificació email"

    var link *url.URL
    link, err := url.Parse(config.C.WebServer.Prefix)
    if err != nil {
        return err
    }
    link.Path += "/emailreg"
    params := url.Values{}
    params.Add("address", address)
    params.Add("code", AuthCode(address,email))
    link.RawQuery = params.Encode()
    linkText := link.String()

    msg := "<h1>Blockchain catalunya</h1><br>Feu click <a href='"+linkText+"''>Aqui</a> per verificar el vostre email"

	e.HTML = []byte(msg)

	return e.SendWithTLS(config.C.SmtpClient.Server,auth,nil)
}

