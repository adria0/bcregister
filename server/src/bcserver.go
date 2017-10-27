package main

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/common"	
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	emailer "github.com/jordan-wright/email"
	"github.com/spf13/viper"
	"encoding/json"
	"crypto/hmac"
	"crypto/sha256"
	"net/smtp"
	"strings"
	"net/url"
	"io/ioutil"
	"log"
	"fmt"
)

type Config struct {
	WwwRoot string
	DataFolder string
	WebPrefix string
	EmailSender string
	EmailAuthCode string
	SmtpServer struct {
		Server string
		User	string
		Password string
		Domain string
	}
}

var C Config



type JsonRpcInMsg struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []interface{} `json:"params"`
	Id      uint `json:"id"`
}

type JsonRpcErrorMsg struct {
	Code uint `json:"code"`
	Message string `json:"message"`
}

type JsonRpcOutMsg struct {
	Jsonrpc string `json:"jsonrpc"`
	Error *JsonRpcErrorMsg `json:"error,omitempty"`
	Data interface{} `json:"data,omitempty"`
	Id uint `json:"id"`
}

var (

	errInternalError = &JsonRpcErrorMsg{
		Code : 999,
		Message : "Error intern",
	}

	errUnknownMethod = &JsonRpcErrorMsg{
		Code : 990,
		Message : "Unknown method",
	}

	errBadMsgFormat = &JsonRpcErrorMsg{
		Code : 991,
		Message : "Bad message format",
	}

	errBadSignature = &JsonRpcErrorMsg{
		Code : 992,
		Message : "Bad signature",
	}

	errUnregistered = &JsonRpcErrorMsg{
		Code : 993,
		Message : "No registrat",
	}

)

const (
	emailAuthCode = "D354151CE23AF56154AE7BB313683"
)

func getEmailAuthCode(address, email string) string {
	mac := hmac.New(sha256.New, []byte(emailAuthCode))
	mac.Write([]byte(address))
	mac.Write([]byte(email))
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum)
}

func sendAuthEmail(address, email string) error {

	auth := smtp.PlainAuth(
		"",
		C.SmtpServer.User,
		C.SmtpServer.Password,
		C.SmtpServer.Domain,
	)

	e := emailer.NewEmail()
	e.Headers.Add("Content-Transfer-Encoding","quoted-printable")
	e.From = C.EmailSender
	e.To = []string{email}
	e.Subject = "Blockchain Catalunya - Verificació email"

    var link *url.URL
    link, err := url.Parse(C.WebPrefix)
    if err != nil {
        return err
    }
    link.Path += "/emailreg"
    params := url.Values{}
    params.Add("address", address)
    params.Add("code", getEmailAuthCode(address,email))
    link.RawQuery = params.Encode()
    linkText := link.String()

    msg := "<h1>Blockchain catalunya</h1><br>Feu click <a href='"+linkText+"''>Aqui</a> per verificar el vostre email"

	e.HTML = []byte(msg)

	return e.Send("authsmtp.blockchaincatalunya.org:587",auth)
}

func verifyMsg(in JsonRpcInMsg) ([]interface{}, string,error) {

	args := in.Params[:len(in.Params)-2]
	address := in.Params[len(in.Params)-2].(string)

	sig,err := hex.DecodeString(in.Params[len(in.Params)-1].(string))
	if err != nil {
		return nil,"",err
	}
	sig[len(sig)-1] -= 27;

	signedData := append(args,in.Method, in.Id)
	data, err := rlp.EncodeToBytes(signedData)
	if err != nil {
		return nil,"",err
	}

	pubKey, err := crypto.Ecrecover(crypto.Keccak256(data), sig)
	if err != nil {
		return nil,"",err
	}

	sigAddress := common.BytesToAddress(common.LeftPadBytes(crypto.Keccak256(pubKey[1:])[12:], 32))
	sigAddressStr := strings.ToLower(sigAddress.String())
	if sigAddressStr != address {
		return nil,"",fmt.Errorf("Signature mismatch")
	}

	return args,address,nil
}

type BcMember struct {
	Address string `json:"address"`
	FirstName string `json:"firstName"`
	SecondName string `json:"secondName"`
	Mode string `json:"mode"`
	Interest string `json:"interest"`
	Email string `json:"email"`
	EmailVerified bool `json:"emailVerified"`
}

type Directory struct {
	DataFolder string
}

func NewDirectory(dataFolder string) *Directory {
	return &Directory{dataFolder}
}

func (d *Directory) Add(member *BcMember) error {

	serialized, err := json.Marshal(member)

	if err != nil {
		return err
	}

	return ioutil.WriteFile("member-"+member.Address,serialized,0666)	
}

func (d *Directory) Update(member *BcMember) error {

	serialized, err := json.Marshal(member)

	if err != nil {
		return err
	}

	return ioutil.WriteFile("member-"+member.Address,serialized,0666)	
}

func (d *Directory) Read(address string) (*BcMember,error) {

	serialized, err := ioutil.ReadFile("member-"+address)
	if err != nil {
		return nil,err
	}

	var member BcMember
	err = json.Unmarshal([]byte(serialized), &member)
	if err!=nil {
		return nil, err
	}

	return &member, nil
}

func dispatchMsg(directory *Directory, address, method string, args []interface{}) (interface{},*JsonRpcErrorMsg) {

	if (method == "bc_register" ) {

		if len(args) != 5 {
			return nil, errInternalError
		}

		firstName := args[0].(string)
		secondName := args[1].(string)
		email := args[2].(string)
		mode := args[3].(string)
		interest := args[4].(string)

		member, err := directory.Read(address)
		if err == nil {
			if member.EmailVerified {
				err = fmt.Errorf("Address already registred")
			}
		}

		err = directory.Add(&BcMember{
			Address : address,
			FirstName : firstName,
			SecondName : secondName,
			Email : email,
			Mode : mode,
			Interest: interest,
		})

		if err!=nil {
			return nil, errInternalError
		}
		
		sendAuthEmail(address,email)
		return nil, nil

	}

	if (method == "bc_auth" ) {

		member, err := directory.Read(address)
		if err != nil {
			return nil, errUnregistered
		}
		return member, nil
	}

	return nil, errUnknownMethod

}

func main() {

	viper.SetConfigType("yaml")
	viper.SetConfigName("bcserver")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("BCSERVER") 
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	if err := viper.Unmarshal(&C); err != nil {
		log.Fatal(err)
	}

	directory := NewDirectory(C.DataFolder)

	r := gin.Default()

    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Token")
        c.Next()
    })

	r.GET("/emailreg", func(c *gin.Context) {
		address := c.Query("address")
		code := c.Query("code")

		member, err := directory.Read(address)
		if err == nil {
			expectedCode := getEmailAuthCode(address,member.Email)
			if expectedCode == code {
				member.EmailVerified = true
				directory.Update(member)
			} else {
				err = fmt.Errorf("Codi incorrecte")
			}
		}

		if err == nil {
			c.String(200, "Correu registrat correctament.")
		} else {
			c.String(200, "No s'ha pogut regisrar el correu.")
		}

	})

	r.Static("/r", C.WwwRoot)

    r.OPTIONS("/*cors", func(c *gin.Context) {
    })

	r.POST("/rpc", func(c *gin.Context) {

		var err error
		var in JsonRpcInMsg
		var rpcErr *JsonRpcErrorMsg
		var retvalue interface{}

		inraw, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Failed reading request body",err)
			return
		}
		
		dec := json.NewDecoder(strings.NewReader(string(inraw)))
		err = dec.Decode(&in)

		if err == nil {
			var address string
			var args []interface{}

			args,address,err = verifyMsg(in)

			if err == nil {
				retvalue, rpcErr = dispatchMsg(directory,address,in.Method,args)
			} else {
				rpcErr = errBadSignature
			}

		} else {
			log.Printf("[%s]",string(inraw))
			rpcErr = errBadMsgFormat
		}

		out := &JsonRpcOutMsg{
			Jsonrpc : "2.0",
			Id : in.Id,
			Error : rpcErr,
			Data : retvalue,
		}

		c.JSON(200, out)

		if rpcErr != nil {
			log.Printf("%#v => %#v [rpcerr=%s] [interr=%v]\n",
				in,out,rpcErr.Message,err)
		} else {
			log.Printf("%#v => %#v\n",in,out)
		}

	})
	r.Run()
}
