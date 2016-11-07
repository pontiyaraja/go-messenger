package core

import (
	"bitbucket.org/tekion/tekionbaas/log"
	"bitbucket.org/tekion/tmessenger/utills"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"sync"
	"time"
)

var clients = struct {
	sync.RWMutex
	m map[string]*client
}{m: make(map[string]*client)}

var rooms = struct {
	sync.RWMutex
	m map[string]*roomRO
}{m: make(map[string]*roomRO)}

var clientsRO = struct {
	sync.RWMutex
	m map[string]*client
}{m: make(map[string]*client)}

type room struct {
	clientMessages   chan *ClientSocketMessage
	serverMessage    chan *ServerAction
	roIssueMessage   chan *ROIssueMessage
	pongMessage      chan *MsgResponse
	broadcastMessage chan *RFID_receive
	join             chan *client
	leave            chan *client
	clients          map[*client]bool
}

type roomRO struct {
	name             string
	roIssueMessage   chan *ROIssueMessage
	pongMessage      chan *MsgResponse
	broadcastMessage chan *RFID_receive
	join             chan *client
	leave            chan *client
	clients          map[*client]bool
}

type RoomRo struct {
	name  string
	rooms map[*roomRO]bool
}

// All clients for this event..for an event many clients might be interested
var eventRooms = struct {
	sync.RWMutex
	m map[string][]*client
}{m: make(map[string][]*client)}

func addClientToEvent(eventId string, c *client) {
	eventRooms.Lock()
	eventRooms.m[eventId] = append(eventRooms.m[eventId], c)
	eventRooms.Unlock()
}

func GetClientsForEvent(eventId string) []*client {
	return eventRooms.m[eventId]
}

func AddClient(clientId string, c *client) {
	clients.Lock()
	clients.m[clientId] = c
	clients.Unlock()
}

func AddRoooms(roomId string, r *roomRO) {
	rooms.Lock()
	rooms.m[roomId] = r
	rooms.Unlock()
}

func GetClient(clientId string) (c *client) {
	clients.RLock()
	c = clients.m[clientId]
	clients.RUnlock()
	return c
}

func GetRoom(roomId string) (r *roomRO) {
	rooms.RLock()
	r = rooms.m[roomId]
	rooms.RUnlock()
	return r
}

func RemoveClient(clientId string) {
	clients.Lock()
	delete(clients.m, clientId)
	clients.Unlock()
}

func RemoveRoom(roomId string) {
	rooms.Lock()
	delete(rooms.m, roomId)
	rooms.Unlock()
}

func AddROClient(clientId string, c *client) {
	clientsRO.Lock()
	clientsRO.m[clientId] = c
	clientsRO.Unlock()
}

func newRoom() *room {
	return &room{
		clientMessages:   make(chan *ClientSocketMessage),
		serverMessage:    make(chan *ServerAction),
		roIssueMessage:   make(chan *ROIssueMessage),
		broadcastMessage: make(chan *RFID_receive),
		pongMessage:      make(chan *MsgResponse),
		join:             make(chan *client),
		leave:            make(chan *client),
		clients:          make(map[*client]bool),
	}
}

func newRoomRO() *roomRO {
	return &roomRO{
		roIssueMessage:   make(chan *ROIssueMessage),
		broadcastMessage: make(chan *RFID_receive),
		pongMessage:      make(chan *MsgResponse),
		join:             make(chan *client),
		leave:            make(chan *client),
		clients:          make(map[*client]bool),
	}
}

func newRoomRo() *RoomRo {
	return &RoomRo{
		rooms:          make(map[*roomRO]bool),
	}
}

func (r *room) run() {
	log.Info("Starting the Lobby Room....\n")
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			msg := MsgResponse{Msg: MESSAGE_USER_FOUND, Data: client.clientId + " just joined in lobby"}
			fmt.Println("MEssage MSG -----> ", msg)
			b, err := json.Marshal(msg)
			if err == nil {
				utills.PublishMessage("service lobby", string(b))
			}
		case client := <-r.leave:
			delete(r.clients, client)
			client.ws.Close()
		case msg := <-r.clientMessages:
			go func() {
				log.Info("Recieved Message in Lobby From Client id", msg.FromClient)
				toClient := GetClient(msg.ToClient)
				if toClient != nil {
					log.Info("Sending message to id", msg.ToClient, msg.D)
					mesgData := &ClientSocketMessage{Msg: msg.Msg, FromClient: msg.FromClient,
						ToClient: msg.ToClient, D: msg.D, TimeStamp: time.Now().UTC()}
					toClient.send <- mesgData
					b, err := json.Marshal(mesgData)
					if err == nil {
						erResp := utills.SetMessageData(getChatKey(msg.FromClient, msg.ToClient), string(b))
						if erResp != nil {
							fmt.Println("Set message error -------------> ", erResp)
							panic(erResp)
						}
					} else {
						panic(err)
					}
				}
			}()
		case sMsg := <-r.serverMessage:
			// This should be replaced with redis pub-subscribe as its going to be expensive operation
			for client := range r.clients {
				Msg := &ClientSocketMessage{Msg: sMsg.Msg, FromClient: client.clientId, D: sMsg.Message}
				client.send <- Msg
			}
		case rfid := <-r.broadcastMessage:
			fmt.Println("In brodcast channel")
			fmt.Println("rfid : ", rfid)
			rfidMsg := RFID_receive{Rfid:rfid.Rfid, Msg:rfid.Msg}
			msg, err := json.Marshal(rfidMsg)
			if err != nil {
				panic(err)
			}
			utills.PublishMessage("service lobby", string(msg))
			//for client := range r.clients {
			//	fmt.Println("Client ")
			//	fmt.Println(client)
			//	rfidMsg := RFID_receive{Rfid:rfid.Rfid, Msg:rfid.Msg}
			//	//rfidMsg := RFID_receive{Rfid: rfid.Rfid, CustomerInfo: rfid.CustomerInfo}
			//	client.broadcastsend <- &rfidMsg
			//}
		case roIssue := <-r.roIssueMessage:
			toClient := GetClient(roIssue.ToClient)
			if roIssue.Msg == MESSAGE_CHAT_RO_ISSUE {
				if toClient != nil {
					mesgData := ROIssueMessage{Msg: roIssue.Msg, D:roIssue.D, FromClient: roIssue.FromClient,
						ToClient: roIssue.ToClient, ROId: roIssue.ROId, TimeStamp: time.Now().UTC()}
					toClient.roSend <- &mesgData
					b, err := json.Marshal(mesgData)
					if err == nil {
						fmt.Println("String value of message ---> ",string(b))
						erResp := utills.SetMessageData(getChatKey(roIssue.FromClient, roIssue.ToClient), string(b))
						if erResp != nil {
							fmt.Println("Set message error -------------> ", erResp)
							panic(erResp)
						}
					} else {
						panic(err)
					}
				}
			}else if roIssue.Msg == MESSAGE_EVENTAPPOINTMENT{
				if toClient != nil {
					msg := &ROIssueMessage{Msg:roIssue.Msg, ToClient:roIssue.ToClient, D:roIssue.D, AppointmentId:roIssue.AppointmentId, TimeStamp:time.Now().UTC()}
					toClient.roSend <- msg
				}
			}else {
				if toClient != nil {
					msg := &ROIssueMessage{Msg:roIssue.Msg, ToClient:roIssue.ToClient, D:roIssue.D, ROId:roIssue.ROId, TimeStamp:time.Now().UTC()}
					toClient.roSend <- msg
				}
			}

		}
	}
}

const (
	socketBufferSize  = 12288
	messageBufferSize = 12288
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// format: /join/{client-id}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Println("here url", req.URL)
	var response interface{}
	json.NewDecoder(req.Body).Decode(&response)
	log.Info(response)
	args := strings.Split(req.URL.Path, "/")
	/*accessToken, err := base64.StdEncoding.DecodeString(req.URL.Query().Get(ACCESS_TOKEN))
	//accessToken := req.URL.Query().Get("accessToken")
	if err != nil {
		fmt.Println("error:", err)
		return
	}*/
	clientId := args[2]
	fmt.Println("Client ID is ", clientId)

	//Authenticating client
	/*message := utills.AuthenticateUser(clientId, string(accessToken))
	if strings.Compare(message, SUCCESS) != 0 {
		w.Header().Set(WEBSOCKER_ERR_MSG,message)
		return
	}*/
	/*<--  Authentication client end  -->*/

	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Debug("ServeHTTP:", err)
		return
	}
	log.Info("Remote address ----------> ", socket.RemoteAddr().String())
	client := newClient(socket, r, clientId)
	r.join <- client
	AddClient(clientId, client)
	msg := constructMessage(MESSAGE_SUCCESS, nil)
	b, err := json.Marshal(msg)
	fmt.Printf("msg : +%v\n", msg)
	SendDataToWS(client.ws, b)
	//notifyClientEntry(client)

	defer func() { r.leave <- client }()
	go client.write()
	go client.writeRO()
	go client.broadcastwrite()
	go client.subscribe("service lobby")
	log.Info("after client write")
	go client.messagePing()
	go client.handleMessage()
	client.read()

	
}

func (r *room) AddEventToNotify(w http.ResponseWriter, req *http.Request) {
	fmt.Println("GS:Entering AddActionToNotifyfunc  ", req.URL.Path)
	args := strings.Split(req.URL.Path, "/")
	actionId := args[2]
	clientId := args[3]
	addClientToEvent(actionId, GetClient(clientId))
}

func SendDataToWS(con *websocket.Conn, str []byte) error {
	return con.WriteMessage(websocket.TextMessage, str)
}

func constructMessage(msgType uint64, data interface{}) ClientResponse {
	var msg ClientResponse
	msg = ClientResponse{Msg: msgType, Data: data}
	return msg
}

func ConstructMsg(msgType int64, data string) MsgResponse {
	var msg MsgResponse
	msg = MsgResponse{Msg: msgType, Data: data}
	return msg
}

func getChatKey(fromClient, toClient string) string {
	key := ""
	if strings.Compare(fromClient, toClient) == -1 {
		key = fromClient + COLON + toClient
	} else {
		key = toClient + COLON + fromClient
	}
	return key
}