package messageQueue

import (
	"bitbucket.org/tekion/tmessenger/utills"
	"fmt"
	"reflect"
	"bitbucket.org/tekion/tmessenger/core"
)

type Queue struct {
	Name   string
	UserId string `json:"userId"`
}

func init() {
	c := core.GetClient("pan")
	ws := reflect.TypeOf(c)
	fmt.Println("WS -------->  ",ws)
	//go getQueue()
}

func AddCalendarQueue(event string) bool  {
	fmt.Println("In add event ------> ", event)
	err := utills.RPush("EventQueue", event)
	if err != nil {
		return false
	}else {
		getQueue()
		return true
	}
	return true
}

func getQueue() {
	err := utills.RPopLPush("EventQueue", "EventQueueTemp")
	if err != nil {
		fmt.Println("Problem in RPopLPush", err)
	}else {
		lrang, err := utills.LRange("EventQueueTemp")
		fmt.Println("Queue ------> ", err)
		fmt.Println("Queue -------> ", lrang)
		if err == nil {
			for i := 0; i < len(lrang); i++ {
				fmt.Println("Queue in for  -------> ", lrang[i])
				ProcessEventMsg(lrang[i])
			}

		}
	}
}

