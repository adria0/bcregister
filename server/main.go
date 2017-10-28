package main

import (
	"fmt"
	"github.com/adriamb/bcdapp/config"
	"github.com/adriamb/bcdapp/db"
	"github.com/adriamb/bcdapp/email"
	"github.com/adriamb/bcdapp/jsonrpc"
	"github.com/adriamb/bcdapp/recaptcha"
	"github.com/gin-gonic/gin"
	"log"
)

var (
	errInternalError = &jsonrpc.ErrorMsg{
		Code:    999,
		Message: "Error intern",
	}

	errUnknownMethod = &jsonrpc.ErrorMsg{
		Code:    990,
		Message: "Unknown method",
	}

	errUnregistered = &jsonrpc.ErrorMsg{
		Code:    993,
		Message: "No registrat",
	}

	errAlreadyRegistered = &jsonrpc.ErrorMsg{
		Code:    994,
		Message: "Adreça ja registrada",
	}
)

func jsonRpcDispatcher(c *gin.Context, address, method string, args []interface{}) (interface{}, *jsonrpc.ErrorMsg) {

	if method == "bc_register" {

		if len(args) != 6 {
			log.Print("*err bad-bc-register-args", len(args))
			return nil, errInternalError
		}

		firstName := args[0].(string)
		secondName := args[1].(string)
		useremail := args[2].(string)
		mode := args[3].(string)
		interest := args[4].(string)
		captcha := args[5].(string)

		err := recaptcha.Verify(config.C.Recaptcha.Key, captcha, c.ClientIP())
		if err != nil {
			log.Print("*err recaptcha-verify", err)
			return nil, errInternalError
		}

		member, err := db.Read(address)
		if err == nil {
			if member.EmailVerified {
				return nil, errAlreadyRegistered
			}
		}

		err = db.Add(&db.Member{
			Address:    address,
			FirstName:  firstName,
			SecondName: secondName,
			Email:      useremail,
			Mode:       mode,
			Interest:   interest,
		})

		if err != nil {
			log.Print("*err db-add", err)
			return nil, errInternalError
		}
		/* err = email.SendAuthEmail(address, useremail)
		if err != nil {
			log.Print("*err db-sendmail", err)
			return nil, errInternalError
		}
		*/
		return nil, nil
	}

	if method == "bc_auth" {

		member, err := db.Read(address)
		if err != nil {
			return nil, errUnregistered
		}
		return member, nil
	}

	return nil, errUnknownMethod

}

func GETVerifyEmail(c *gin.Context) {

	address := c.Query("address")
	code := c.Query("code")

	member, err := db.Read(address)
	if err == nil {
		expectedCode := email.AuthCode(address, member.Email)
		if expectedCode == code {
			member.EmailVerified = true
			db.Update(member)
		} else {
			err = fmt.Errorf("Codi incorrecte")
		}
	}

	if err == nil {
		c.String(200, "Correu registrat correctament.")
	} else {
		c.String(200, "No s'ha pogut regisrar el correu.")
	}

}

func main() {

	r := gin.Default()

	jsonrpc.SetDispatcher(jsonRpcDispatcher)
	r.POST("/rpc", jsonrpc.Handle)

	r.GET("/emailreg", GETVerifyEmail)
	r.Static("/r", config.C.WebServer.WwwRoot)

	r.RunTLS(config.C.WebServer.Bind, config.C.WebServer.CertFile, config.C.WebServer.KeyFile)
}
