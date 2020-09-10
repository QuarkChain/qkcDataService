package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"net/http"
)

type ReturnMsg struct {
	Code CodeStatus  `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type MainController struct {
	beego.Controller
}

func EncodeRet(ret []byte) interface{} {
	if ret == nil {
		return nil
	}
	return json.RawMessage(string(ret))
}

func (c *MainController) Get() {
	txHash := c.GetString("txHash")
	code := Succ
	fmt.Println("READY GET", "txHash", txHash)
	ret, err := SDK.GetTransactionById(txHash)
	if err != nil {
		code = Failed
	}

	c.Data["json"] = &ReturnMsg{
		Code: code,
		Msg:  GetMsgFromCode(code, err),
		Data: EncodeRet(ret),
	}

	c.ServeJSON()
	fmt.Println("END GET", string(ret))
}

func (c *MainController) Post() {
	fmt.Println("READY POST", "len", len(c.Ctx.Input.RequestBody), string(c.Ctx.Input.RequestBody))
	var (
		code = Succ
		hash = ""
		err  error
	)
	if len(c.Ctx.Input.RequestBody) > MaxPostLen {
		code = Failed
		err = fmt.Errorf("len(requestBody)%d should small than %v", len(c.Ctx.Input.RequestBody), MaxPostLen)
		goto end
	}

	if json.Valid(c.Ctx.Input.RequestBody) == false {
		code = Failed
		err = fmt.Errorf("bad json format")
		goto end
	}

	hash, err = SDK.SendFormData(SDK.GetNonce(), c.Ctx.Input.RequestBody)
	if err != nil {
		code = Failed
	}
	if hash == "0x000000000000000000000000000000000000000000000000000000000000000000000000" {
		code = Failed
		err = errors.New("tx is failed")
	}

end:
	if code == Failed {
		SDK.resetNonce()
	}
	c.Data["json"] = &ReturnMsg{
		Code: code,
		Msg:  GetMsgFromCode(code, err),
		Data: hash,
	}
	c.ServeJSON()
	fmt.Println("END POST", "status", c.Data["json"])
}

func (c *MainController) ServeJSON() {
	c.Ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
	var err error
	data := c.Data["json"]
	encoder := json.NewEncoder(c.Ctx.Output.Context.ResponseWriter)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(data)
	if err != nil {
		http.Error(c.Ctx.Output.Context.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}
