package controllers

import "fmt"

type CodeStatus int

const (
	Succ CodeStatus = iota
	Failed
)

func GetMsgFromCode(code CodeStatus, err error) string {
	switch code {
	case Succ:
		return "successfully"
	case Failed:
		return fmt.Sprintf("failed:%v", err.Error())
	default:
		panic(fmt.Errorf("code not support %v", code))
	}
}
