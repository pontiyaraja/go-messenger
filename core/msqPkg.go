package core

import (
	"bitbucket.org/tekion/tmessenger/utills"
	"bitbucket.org/tekion/tekionbaas/log"
	"fmt"
)

var roRoom = newRoomRo()

func StartupExistingChannels() {
	for i := 0; i < 1; i++ {
		StartExistingSubscribers("CHANNELS")
	}
}

func AddROClientM(clientId, roomId string)  {
	rRo := GetRoom(roomId)
	c := GetClient(clientId)
	AddROClient(clientId, c)
	rRo.clients[c] = true
}

func AddRORooms(roomId string){
	rRo := GetRoom(roomId)
	if rRo == nil {
		rRo = newRoomRO()
		AddRoooms(roomId, rRo)
		roRoom.rooms[rRo] = true
	}
}

func AddChannel(clientId, roId string){
	chExist, cheror := utills.HExists("CHANNELS", roId)
	if !chExist && cheror == nil {
		updated, err := utills.Hset("CHANNELS", roId, roId)
		if !updated && err != nil {
			log.Info("Error adding channels.....  ", err)
		}
	}
	exist, eror := utills.HExists(roId, clientId)
	if !exist && eror == nil {
		updated, err := utills.Hset(roId, clientId, roId)
		if updated && err == nil {
			c := GetClient(clientId)
			go c.subscribe(roId)
		}
	}
}

func StartExistingSubscribers(key string)  {
	chanMap, err := utills.HGetAll("CHANNELS")
	if err == nil {
		for field := range chanMap {
			exist, eror := utills.HExists(field, key)
			if exist && eror == nil {
				c := GetClient(key)
				if c != nil{
					go c.subscribe(field)
					break
				}
			}

		}
	}
}

func AddExistingSubscribers(key string)  {
	chanMap, err := utills.HGetAll(key)
	if err == nil {
		for field := range chanMap {
			fmt.Println("Field --> ", field)
			keyMap, eror := utills.HGetAll(field)
			if eror == nil {
				for client, roId := range keyMap {
					fmt.Println("Client --> ", client)
					c := GetClient(client)
					if c != nil{
						go c.subscribe(roId)
					}
				}

			}
		}
	}
}


