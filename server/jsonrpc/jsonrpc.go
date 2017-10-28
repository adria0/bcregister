package jsonrpc

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/common"	
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"strings"
	"log"
	"encoding/hex"
)

type InMsg struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []interface{} `json:"params"`
	Id      uint `json:"id"`
}

type ErrorMsg struct {
	Code uint `json:"code"`
	Message string `json:"message"`
}

type OutMsg struct {
	Jsonrpc string `json:"jsonrpc"`
	Error *ErrorMsg `json:"error,omitempty"`
	Data interface{} `json:"data,omitempty"`
	Id uint `json:"id"`
}

var (

	errBadMsgFormat = &ErrorMsg{
		Code : 10001,
		Message : "Bad message format",
	}

	errBadSignature = &ErrorMsg{
		Code : 10002,
		Message : "Bad signature",
	}

	dispatcher func (c *gin.Context, address, method string, args []interface{}) (interface{},*ErrorMsg)
)

func verifyMsg(in InMsg) ([]interface{}, string,error) {

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

func SetDispatcher(d func (c *gin.Context, address, method string, args []interface{}) (interface{},*ErrorMsg)) {
	dispatcher = d
}

func Handle(
	c *gin.Context,
	) {

	var err error
	var in InMsg
	var rpcErr *ErrorMsg
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
			retvalue, rpcErr = dispatcher(c,address,in.Method,args)
		} else {
			rpcErr = errBadSignature
		}

	} else {
		log.Printf("[%s]",string(inraw))
		rpcErr = errBadMsgFormat
	}

	out := &OutMsg{
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

}


