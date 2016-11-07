package core

import (
	"bitbucket.org/tekion/tekionbaas/log"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"bitbucket.org/tekion/tmessenger/utills"
)

var AuthToken string
var CustomerInfoUrl = "http://api.tekion.xyz/customer"

var serverUrl = "http://api.tekion.xyz/customer"

func GetAuthToken() {
	url := "http://api.tekion.xyz/login"
	var jsonStr = []byte(`{"username":"tmessenger@tekion.com","password":"Tekion123"}`)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonStr))
	client := &http.Client{}
	resp, _ := client.Do(req)
	type Tmp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	response := Tmp{}
	json.NewDecoder(resp.Body).Decode(&response)
	AuthToken = response.Data.AccessToken
}

// format: /notify/{client-action}/rfid - customer arrival - 6

// format: /notify/{client-action}/roid/{user-id} - ro assignment to sa - 10

// format: /notify/{client-action}/roid/[user-id} - ro pending - 29

// format: /notify/{client-action} - get online users - 30

func actionNotifyHandler(rm *room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		args := strings.Split(r.URL.Path, "/")
		fmt.Println(args)
		action, err := strconv.Atoi(args[2])
		fmt.Println(action)
		if err != nil {
			log.Info("no proper value passed in notify", err)
		} else {
			switch action {
			case TOPIC_RFID_RECEIVE:
				rfid := args[3]
				decoder := json.NewDecoder(r.Body)
				res := RFID_receive{}
				err := decoder.Decode(&res)
				if err != nil {
					fmt.Sprintf("%s", err)
				}
				fmt.Println("switch = ", rfid)
				//customer := getClientInfo(rfid)
				msg := &RFID_receive{Rfid:rfid, Msg:int64(action)}
				//msg := &RFID_receive{Rfid: rfid, CustomerInfo: customer.Data}
				fmt.Println("rfid")
				fmt.Print(msg.Rfid)
				rm.broadcastMessage <- msg

			case MESSAGE_RO_ASSIGNED:
				roid := args[3]
				userId := args[4]
				AddChannel(userId, roid)
				decoder := json.NewDecoder(r.Body)
				res := ROMessage{}
				err := decoder.Decode(&res)
				if err != nil {
					fmt.Sprintf("%s", err)
				}
				roMsg := "Repair Order "+ roid +" assigned to "+userId
				fmt.Println("MEsage --------> ", roMsg)
				msg := ROMessage{Msg:int64(action), ToClient:userId, D:roMsg, ROId:roid}
				b, err := json.Marshal(msg)
				if err == nil {
					utills.PublishMessage(roid, string(b))
				}
				fmt.Println("msg in ro : ", msg)
				//rm.roIssueMessage <- msg

			case MESSAGE_RO_CREATED:
				roid := args[3]
				userId := args[4]
				AddChannel(userId, roid)
				decoder := json.NewDecoder(r.Body)
				res := ROIssueMessage{}
				err := decoder.Decode(&res)
				if err != nil {
					fmt.Sprintf("%s", err)
				}
				roMsg := "A new repair order "+ roid +" created"
				AddChannel(userId, roid)
				utills.PublishMessage(roid, roMsg)
				msg := &ROIssueMessage{Msg:int64(action), ToClient:userId, D:roMsg, ROId:roid}
				fmt.Println("msg in ro : ", msg)
				rm.roIssueMessage <- msg

			case MESSAGE_RO_ASSIGNED_PARTS:
				roid := args[3]
				userId := args[4]
				decoder := json.NewDecoder(r.Body)
				res := ROIssueMessage{}
				err := decoder.Decode(&res)
				if err != nil {
					fmt.Sprintf("%s", err)
				}
				roMsg := "Repair Order "+ roid +" assigned to parts"
				msg := &ROIssueMessage{Msg:int64(action), ToClient:userId, D:roMsg, ROId:roid}
				fmt.Println("msg in ro : ", msg)
				rm.roIssueMessage <- msg

			case MESSAGE_RO_ASSIGNED_INVOICE:
				roid := args[3]
				userId := args[4]
				decoder := json.NewDecoder(r.Body)
				res := ROIssueMessage{}
				err := decoder.Decode(&res)
				if err != nil {
					fmt.Sprintf("%s", err)
				}
				addClientToEvent("customerArrival", GetClient(userId))
				roMsg := "Repair Order "+ roid +" assigned to invoice"
				msg := &ROIssueMessage{Msg:int64(action), ToClient:userId, D:roMsg, ROId:roid}
				fmt.Println("msg in ro : ", msg)
				rm.roIssueMessage <- msg

			case MESSAGE_EVENTAPPOINTMENT:
				eventId := args[3]
				userId := args[4]
				decoder := json.NewDecoder(r.Body)
				res := ROIssueMessage{}
				err := decoder.Decode(&res)
				if err != nil {
					fmt.Sprintf("%s", err)
				}
				roMsg := "A new appointment "+ eventId +" assigned to "+userId
				msg := &ROIssueMessage{Msg:int64(action), ToClient:userId, D:roMsg, AppointmentId:eventId}
				rm.roIssueMessage <- msg

			case MESSAGE_RO_PENDING:
				roid := args[3]
				userId := args[4]
				decoder := json.NewDecoder(r.Body)
				res := ROIssueMessage{}
				err := decoder.Decode(&res)
				if err != nil {
					fmt.Sprintf("%s", err)
				}
				romsg := "Repair Order " + roid + " is pending. Would you like to finish it?"
				msg := &ROIssueMessage{Msg:int64(action), ToClient:userId, D:romsg, ROId:roid}
				fmt.Println("msg in ro pending: ", msg)
				rm.roIssueMessage <- msg

			case MESSAGE_SERVER_ACTION_PUSH:
				log.Info("in server Action Push")
				msg := &ServerAction{Msg: int64(action), Message: args[3]}
				rm.serverMessage <- msg

			case MESSAGE_AVAILABLE_USERS:
				log.Info("fetching available users")
				username := r.FormValue("u")
				if username != "" {
					add_online_username(username)
				}

			default:
				log.Info(w, "action %s not supported", action)
			}
		}
	}
}

// format: rfid/info

func getClientInfo(rfid string) CustomerInfo {
	customerRfid := RFID_receive{Rfid: rfid}
	r, _ := json.Marshal(customerRfid)
	Bsseurl := serverUrl + "/" + rfid
	request, _ := http.NewRequest(GET, Bsseurl, bytes.NewBuffer(r))
	//request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set(CONTENT_TYPE, CONTENT_VALUE)
	request.Header.Set(TEKION_API_TOKEN, AuthToken)
	client := http.Client{}
	response, _ := client.Do(request)
	customerName := CustomerInfo{}
	if response.StatusCode != http.StatusOK {
		fmt.Print(response)
		customerName = CustomerInfo{Data: UserD{Fname: "new", Lname: "new"}}
	} else {
		decoder := json.NewDecoder(response.Body)
		err := decoder.Decode(&customerName)
		if err != nil {
			fmt.Println("%s", err)
		}
	}
	return customerName

}

func add_online_username(u string) {

}

func Start() {
	serviceLobby := newRoom()
	http.Handle("/join/", serviceLobby)
	go StartupExistingChannels()
	go serviceLobby.run()
	http.HandleFunc("/notify/", actionNotifyHandler(serviceLobby))
	log.Info("Starting messenger server on 8091")
	if err := http.ListenAndServe(":8091", nil); err != nil {
		log.Debug("ListenAndServe:", err)
	}
}
