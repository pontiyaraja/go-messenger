package core

import (
	"bitbucket.org/tekion/tekionbaas/log"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
	"bitbucket.org/tekion/tmessenger/utills"
	redis "gopkg.in/redis.v4"
)

type client struct {
	clientId      string
	ws            *websocket.Conn
	send          chan *ClientSocketMessage
	receive       chan *ClientSocketMessage
	roSend        chan *ROIssueMessage
	roReceive     chan *ROIssueMessage
	broadcastsend chan *RFID_receive
	hbreceive     chan *MsgResponse
	room          *room
	connected     bool
}

func (c *client) read() {
	var msg = ClientSocketMessage{}
	var roMsg = ROIssueMessage{}
	for c.connected {
		if msgType, byt, eror := c.ws.ReadMessage(); eror == nil {
		fmt.Println("-------->  ", msgType)
			if err := json.Unmarshal(byt, &msg); err == nil && msg.Msg == MESSAGE_CHAT {
				log.Info("the message from client read ", msg.FromClient, "   is", msg.D)
				c.receive <- &msg
			} else if err != nil {
				fmt.Printf("Read Error = %+v\n", err.Error())
				c.connected = false
				break
			} else {
				if err1 := json.Unmarshal(byt, &roMsg); err1 == nil {
					log.Info("the ROmessage from client read ", roMsg.FromClient, "   is", roMsg.D)
					c.roReceive <- &roMsg
				} else if err1 != nil {
					fmt.Printf("Read Error = %+v\n", err.Error())
					c.connected = false
					break
				}
			}
		}else {
			fmt.Println(eror)
			c.connected = false
			break
		}
	}
	c.ws.Close()
}

func (c *client) write() {
	fmt.Println("Write in client")
	for c.connected {
		for msg := range c.send {
			fmt.Println(msg)
			if err := c.ws.WriteJSON(msg); err != nil {
				c.connected = false
				fmt.Printf("Write Error = %+v\n", err.Error())
				break
			}
		}
	}
	c.ws.Close()
}

func (c *client) writeRO() {
	fmt.Println("Write in client")
	for c.connected {
		for msg := range c.roSend {
			fmt.Println(msg)
			if err := c.ws.WriteJSON(msg); err != nil {
				c.connected = false
				fmt.Printf("Write Error = %+v\n", err.Error())
				break
			}
		}
	}
	c.ws.Close()
}

//func (c *client)

func (c *client) broadcastwrite() {
	fmt.Println("broadcast Write in client")

	for c.connected {
		for msg := range c.broadcastsend {
			fmt.Println(msg)
			if err := c.ws.WriteJSON(msg); err != nil {
				fmt.Printf("BWrite Error = %+v\n", err.Error())
				c.connected = false
				break
			}
		}

	}
	c.ws.Close()
}

func (c *client) messagePing() {
	for c.connected {
		/*timeout := time.After(30 * time.Second)
		tick := time.Tick(5000 * time.Millisecond)*/
		time.Sleep(15 * time.Second)
		msgData := ConstructMsg(MESSAGE_SOCKET_PING, PING)
		b, err := json.Marshal(msgData)
		fmt.Println("b : ", b)
		if err == nil {
			SendDataToWS(c.ws, b)
		} else {
			fmt.Printf("Error LOL, %+v", err.Error())
		}
		/*for {
			select {
			case <-timeout:
				fmt.Println("in timeout")
				c.ws.Close()
				delete(c.room.clients, c)
				clientExit := c.clientId + " exit from Lobby "
				c.room.serverMessage <- &ServerAction{Msg: MESSAGE_USER_EXIT, Message: clientExit}
				return
			case pongresponse := <-c.hbreceive:
				fmt.Println("in hbreceive")
				SendDataToWS(c.ws, b)
				if pongresponse.Data == "pong" {
					fmt.Println("in pong")
					SendDataToWS(c.ws, b)
				}
			case <-tick:
				fmt.Println("in tick")
			}
		}*/
	}
}

func newClient(socket *websocket.Conn, r *room, cId string) *client {
	return &client{
		ws:            socket,
		send:          make(chan *ClientSocketMessage),
		receive:       make(chan *ClientSocketMessage),
		roSend:        make(chan *ROIssueMessage),
		roReceive:     make(chan *ROIssueMessage),
		broadcastsend: make(chan *RFID_receive),
		room:          r,
		clientId:      cId,
		connected:     true,
	}
}

func (cl *client) subscribe(channel string) {
	//utills.Init()
	pubsub := utills.SubcribeChannel(channel)
	for {
		msg, err:= pubsub.Receive();
		if err != nil {
			log.Info(err)
			break
		}
		switch redisPacket := msg.(type) {
		case redis.Message:
			redisMesg := RedisMsg{}
			strB, mserr := json.Marshal(msg)
			if mserr == nil{
				unmarhal_errr := json.Unmarshal(strB, &redisMesg)
				if unmarhal_errr != nil {
					fmt.Println("Error in marshaling ...........", unmarhal_errr)
				}
				b, err := json.Marshal(redisMesg.Payload)
				if err == nil {
					SendDataToWS(cl.ws, b)
				}
			}
			log.Info("Message Recieved")
		default:
			redisMesg := RedisMsg{}
			strB, mserr := json.Marshal(msg)
			log.Info("Message Recieved",msg)
			if mserr == nil{
				unmarhal_errr := json.Unmarshal(strB, &redisMesg)
				if unmarhal_errr != nil {
					fmt.Println("Error in marshaling ...........", unmarhal_errr)
				}
				b, err := json.Marshal(redisMesg.Payload)
				log.Info("Message Recieved",redisMesg.Payload)
				if err == nil {
					SendDataToWS(cl.ws, b)
				}
			}
			log.Info("default")
			log.Info(redisPacket)
		}
	}

}

func (cl *client) handleMessage() {

	for {
		select {

		case msg := <-cl.receive:
			switch msg.Msg {
			case MESSAGE_CHAT:

				cl.room.clientMessages <- msg

			case ONLINE_USERS:
				for client := range cl.room.clients {
					fmt.Println("client : ", client.clientId)
					b, err := json.Marshal(client.clientId)
					if err == nil {
						SendDataToWS(client.ws, b)
					}
				}

			case CHECK_AVAILABILITY:
				log.Info("check availability")
				msgData := ConstructMsg(CHECK_AVAILABILITY, NOT_AVAILABLE)
				client := GetClient(msg.ToClient)
				if client == nil {
					msgData = ConstructMsg(CHECK_AVAILABILITY, NOT_AVAILABLE)
					b, err := json.Marshal(msgData)
					if err == nil {
						SendDataToWS(cl.ws, b)
					}
					break
				} else {
					for c := range cl.room.clients {
						if c.clientId == client.clientId {
							msgData = ConstructMsg(CHECK_AVAILABILITY, AVAILABLE)
						}
					}
				}
				b, err := json.Marshal(msgData)
				if err == nil {
					SendDataToWS(cl.ws, b)
				}
			case REDIS_SUBSCRIBE:

			}
		case msg := <-cl.roReceive:
			switch msg.Msg {
			case MESSAGE_CHAT_RO_ISSUE:
				cl.room.roIssueMessage <- msg

			case MESSAGE_ECHO:
				msgData := ConstructMsg(MESSAGE_ECHO, msg.D)
				b, err := json.Marshal(msgData)
				if err == nil {
					SendDataToWS(cl.ws, b)
				}

			/*case MESSAGE_SOCKET_PONG:
				msgData := constructMsg(MESSAGE_SOCKET_PING, PING)
				b, err := json.Marshal(msgData)
				if err == nil {
					SendDataToWS(cl.ws, b)
				} else {
					fmt.Printf("Error LOL, %+v", err.Error())
				}*/
			}
		}
	}

}
