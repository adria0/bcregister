package main

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/common"	
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"strings"
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


func main() {

	r := gin.Default()
	r.POST("/rpc", func(c *gin.Context) {

		var in JsonRpcInMsg
		dec := json.NewDecoder(c.Request.Body)
		err := dec.Decode(&in)
		if err != nil {
			log.Fatal(err)
		}

		args,address,err := verifyMsg(in)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s\n",address)

		out := &JsonRpcOutMsg{
			Jsonrpc : "2.0",
			Id : in.Id,
		}

		c.JSON(200, out)
	})
	r.Run()
}
