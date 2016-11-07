package core

import (
	"net/http"
	"sync"
	"time"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

// temp hack
type ServerAction struct {
	Msg     int64  `json:"msg, string, omitempty"`
	Message string `json:"msg, omitempty"`
}

type RFID_receive struct {
	Msg     int64  `json:"msg, string, omitempty"`
	Rfid         string `json:"rfid"`
	//CustomerInfo UserD  `json:"CustomerInfo"`
}

type CustomerInfo struct {
	Data UserD `json:"data"`
}

type UserD struct {
	Fname string `json:"Fname"`
	Lname string `json:"Lname"`
}

type ClientSocketMessage struct {
	Msg        int64     `json:"msg,omitempty"`
	FromClient string    `json:"fromClient,omitempty"`
	ToClient   string    `json:"toClient,omitempty"`
	D          string    `json:"d,omitempty"`
	TimeStamp  time.Time `json:"timeStamp,omitempty"`
	//D          map[string]interface{} `json:"d"`
}

type ROIssueMessage struct {
	Msg        int64     `json:"msg,omitempty"`
	FromClient string    `json:"fromClient,omitempty"`
	ToClient   string    `json:"toClient,omitempty"`
	D          string    `json:"d,omitempty"`
	ROId       string    `json:"roid,omitempty"`
	AppointmentId string `json:"appointmentId,omitempty"`
	TimeStamp  time.Time `json:"timeStamp,omitempty"`
}

type ClientResponse struct {
	Msg  uint64      `json:"msg"`
	Data interface{} `json:"data"`
}

type PingMessage struct {
	Msg  uint64      `json:"msg"`
	Data interface{} `json:"data"`
}

type MsgResponse struct {
	Msg  int64  `json:"msg"`
	Data string `json:"data"`
}

type User struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	ClientUID   string `json:"clientUID"`
	CreatedTime string `json:"ct"`
	lock        sync.Mutex
}

type OnlineUsers struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	ClientUID   string `json:"clientUID"`
	LastTime    time.Time `json:"lastTime"`
}

//{"Channel":"ROID1234","Pattern":"","Payload":"Repair Order ROID1234 assigned to pan"}

type RedisMsg struct {
	Channel   string `json:"Channel"`
	Pattern   string `json:"Pattern"`
	Payload   string `json:"Payload"`
}

type ROMessage struct {
	Msg        int64     `json:"msg,omitempty"`
	FromClient string    `json:"fromClient,omitempty"`
	ToClient   string    `json:"toClient,omitempty"`
	D          string    `json:"d,omitempty"`
	ROId       string    `json:"roid,omitempty"`
	AppointmentId string `json:"appointmentId,omitempty"`
	TimeStamp  time.Time `json:"timeStamp,omitempty"`
}

const PING string = "ping"
const PONG string = "pong"
const ACCESS_TOKEN string = "accessToken"
const AVAILABLE string = "available"
const NOT_AVAILABLE string = "not available"
const COLON string = ":"
const TEKION_API_TOKEN string = "Tekion-Api-Token"
const CONTENT_TYPE = "Content-Type"
const CONTENT_VALUE = "application/json"
const SUCCESS = "success"
const WEBSOCKER_ERR_MSG = "Websocket-ErrorMsg"
const GET = "GET"

const (
	MESSAGE_SUCCESS            = iota //0
	MESSAGE_ERROR                     // 1
	MESSAGE_NO_USER                   // 2
	MESSAGE_USER_FOUND                // 3
	MESSAGE_USER_CREATED              // 4
	MESSAGE_USER_EXISTS               // 5
	TOPIC_RFID_RECEIVE                // 6
	MESSAGE_SERVER_ACTION_PUSH        // 7
	MESSAGE_CHAT                      // 8
	MESSAGE_CHAT_RO_ISSUE             // 9
	MESSAGE_RO_ASSIGNED
	MESSAGE_TECHNICIAN
	MESSAGE_SOCKET_MESSSAGE
	MESSAGE_SOCKET_MESSSAGE_SUCCESS
	MESSAGE_SOCKET_MESSSAGE_FAILURE
	MESSAGE_SOCKET_OPEN
	MESSAGE_SOCKET_PONG
	MESSAGE_SOCKET_ERROR
	MESSAGE_SOCKET_CLOSE
	MESSAGE_SOCKET_PING
	ONLINE_USERS
	CHECK_AVAILABILITY
	MESSAGE_USER_EXIT
	MESSAGE_USER_ADD
	MESSAGE_ECHO                      //24
	MESSAGE_RO_CREATED
	MESSAGE_RO_ASSIGNED_PARTS
	MESSAGE_RO_ASSIGNED_INVOICE
	MESSAGE_EVENTAPPOINTMENT
	MESSAGE_RO_PENDING          	//29
	MESSAGE_AVAILABLE_USERS		//30
	REDIS_SUBSCRIBE			//31
)
