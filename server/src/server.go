package main

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/common"	
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"strings"
	"io/ioutil"
	"log"
	"fmt"
)

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

)

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
	FirstName string `json:"firstName"`
	SecondName string `json:"secondName"`
	Email string `json:"email"`
}

func dispatchMsg(address, method string, args []interface{}) (interface{},*JsonRpcErrorMsg) {

	if (method == "bc_register" ) {

		serialized, err := json.Marshal(&BcMember{
			FirstName : args[0].(string),
			SecondName : args[1].(string),
			Email :  args[2].(string),
		})
		if err!=nil {
			return nil, errInternalError
		}

		ioutil.WriteFile("member-"+address,serialized,0666)

		return nil, nil

	}

	if (method == "bc_auth" ) {

		serialized, err := ioutil.ReadFile("member-"+address)
		if err != nil {
			return nil, &JsonRpcErrorMsg{
				Code : 1,
				Message : "No registrat",
			}
		}

		var member BcMember
		err = json.Unmarshal([]byte(serialized), &member)
		if err!=nil {
			return nil, errInternalError
		}

		return member, nil

	}

	return nil, errUnknownMethod

}

func main() {

	r := gin.Default()

    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Token")
        c.Next()
    })

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
				retvalue, rpcErr = dispatchMsg(address,in.Method,args)
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
