package messageQueue

import (
	"reflect"
	"encoding/json"
	"fmt"
)

type queue struct {
	Msg    uint64 `json:"msg"`
	ToClient string `json:"toClient"`
	D string `json:"d"`
	RoId   string `json:"roId"`
	TimeStamp string `json:"timeStamp"`
}

func ProcessEventMsg(msg string)  {
	msgQ := queue{}
	msg = reflect.ValueOf(msg).String();
	unmarhal_srrr := json.Unmarshal([]byte(msg), &msgQ)
	fmt.Println("message Queue -------->  ", msgQ.RoId)
	/*client := core.GetClient(msgQ.UserId)
	fmt.Println(client)*/

	if unmarhal_srrr != nil {
		fmt.Println("Error in marshaling ...........", unmarhal_srrr)
		//return errr
	}else {
		fmt.Println("Success -------------> ", msgQ)
	}
}



