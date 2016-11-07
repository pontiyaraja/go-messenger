package utills

import (
	"bitbucket.org/tekion/tekionbaas/log"
	"bitbucket.org/tekion/tmessenger/dao"
	"encoding/json"
	"fmt"
	redis "gopkg.in/redis.v4"
	"strconv"
	"time"
)

type ChatMsg struct {
	Msg        int64     `json:"msg"`
	FromClient string    `json:"fromClient"`
	ToClient   string    `json:"toClient"`
	D          string    `json:"d"`
	ROId       string    `json:"roId"`
	TimeStamp  time.Time `json:"timeStamp"`
}

type MongoData struct {
	MongoArr []dao.ChatMessage
}

var redisClient *redis.Client

func init() {
	redisClient = connectRedis(loadRedisOptions())
	dao.ConnectMongo()
	go updateMessage()
}

func GetRedisClient() *redis.Client {
	return redisClient
}

func connectRedis(options *redis.Options) *redis.Client {
	return redis.NewClient(options)
}

func loadRedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     "10.0.0.167" + ":" + "6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}
}

func Ping() {
	pong, err := redisClient.Ping().Result()
	fmt.Println(pong, err)

}

func Set(key string, value string) error {
	var status *redis.StatusCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.Set(key, value, 0)
		return nil
	})
	if err != nil {
		return err
	}else{
		return status.Err()
	}

}

func SetEx(key string, value string, expireSec int32) {
	var status *redis.BoolCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.SetNX(key, value, time.Duration(expireSec)*time.Second)
		return nil
	})
	if status.Err() != nil && status.Val() {
		log.Info("Redis value inserted => ", key)
	} else {
		log.Error("Redis Error", err)
	}
}

func Get(key string) (string, error) {
	/*var status *redis.StringCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.Get(key)

		return nil
	})
	if err != nil {
		return "", err
	}
	if status.Err() != nil {
		d, e := status.Result()
		return d, e
	}
	return "",err*/
	pipeClient := redisClient.Pipeline();
	r := pipeClient.Get(key)
	_, err := pipeClient.Exec()
	if err == nil {
		d, e := r.Result()
		fmt.Println("Data ---------> ", d, e)
		return d, e
	}
	fmt.Println("Data ---------> ",err)
	return "",err
}

func RPush(listName string, serializedObj string) error {
	var status *redis.IntCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.RPush(listName, serializedObj)
		return nil
	})
	if err != nil {
		return err
	}
	if status.Err() != nil {
		return err
	}
	return nil
}

func RPopLPush(key, newLkey string) error {
	var status *redis.StringCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.RPopLPush(key, newLkey)
		return nil
	})
	if err != nil {
		return err
	}
	if status.Err() != nil {
		return status.Err()
	}
	return nil
}

func SAdd(key string, serializedObj string) error {
	var status *redis.IntCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.SAdd(key, serializedObj)
		return nil
	})
	if err != nil {
		return err
	}
	if status.Err() != nil {
		return status.Err()
	}
	return nil
}

func SMembers(key string) ([]string, error) {
	var status *redis.StringSliceCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.SMembers(key)
		return nil
	})
	if err != nil{
		return nil, err
	}
	d, _ := status.Result()
	return d, status.Err()
}

func Del(key string) error {
	var status *redis.IntCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.Del(key)
		return nil
	})
	if err != nil{
		return err
	}
	return status.Err()
}

func LPush(key, value string) error {
	var status *redis.IntCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.LPush(key, value)
		return nil
	})
	if err != nil{
		return err
	}
	return status.Err()
}

func AddKey(key string) error {
	var resData *redis.IntCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status, eror := Get(key + KEY_COUNT)
		if eror == nil {
			i, err := strconv.ParseFloat(status, 64)
			if err != nil {
				return err
			}
			member := redis.Z{Score: i, Member: key}
			resData = pipe.ZAdd(CHAT_KEY, member)
		}else{
			panic(eror)
		}
		return nil
	})
	if err != nil{
		return err
	}
	return resData.Err()
}

func setNX(key string) error {
	var status *redis.IntCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		keyRes := pipe.SetNX(key, 1, 0)
		if keyRes.Err() == nil {
			status = pipe.Incr(key)
		}
		return nil
	})
	if err != nil{
		return err
	}
	return status.Err()
}

func ZRange(key string) ([]string, error) {
	var status *redis.StringSliceCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.ZRange(key, 0, -1)
		return nil
	})
	if err != nil{
		return nil, err
	}
	if status.Err() != nil {
		return nil, status.Err()
	}
	return status.Val(), nil
}

func LRange(key string) ([]string, error) {
	var status *redis.StringSliceCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.LRange(key, 0, -1)
		return nil
	})
	if err != nil{
		return nil, err
	}
	if status.Err() != nil {
		return nil, status.Err()
	}
	return status.Val(), nil
}

func DelKey(key string) error {
	var status *redis.IntCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.Del(key)
		return nil
	})
	if err != nil{
		return err
	}
	return status.Err()
}

func HExists(key, field string) (bool, error){
	var status *redis.BoolCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.HExists(key, field)
		return nil
	})
	if err != nil{
		return false, err
	}
	return status.Val(), status.Err()
}

func Hset(key, field, value string) (bool, error){
	var status *redis.BoolCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.HSet(key, field, value)
		return nil
	})
	if err != nil{
		return false, err
	}
	return status.Val(), status.Err()
}

func HGetAll(key string) (map[string]string, error) {
	var status *redis.StringStringMapCmd
	_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		status = pipe.HGetAll(key)
		return nil
	})
	if err != nil{
		return nil, err
	}
	return status.Val(), status.Err()
}

func SubcribeChannel(channel string) *redis.PubSub {
	pubsub, err := redisClient.Subscribe(channel)
	if err != nil {
		panic(err)
	}
	time.Sleep(100)
	//defer pubsub.Close()
	return pubsub
}

func PublishMessage(channel string, msg string) {
	fmt.Println("In publish message ---> ", msg)
	n, err := redisClient.Publish(channel, msg).Result()
	fmt.Println(n)
	if err != nil {
		panic(err)
	}
}


func ReceiveMessage(pubSub *redis.PubSub) {
	log.Info("worked")
	for{
		msg, err := pubSub.Receive();
		if err != nil {
			log.Info(err)
			break
		}
		switch redisPacket := msg.(type) {
		case redis.Message:
			log.Info("Message Recieved")
		default:
			log.Info("default")
			log.Info(redisPacket)
		}
	}
}
func SetMessageData(key string, mesg string) error {
	/*_, err := redisClient.Pipelined(func(pipe *redis.Pipeline) error {
		//incr = pipe.Set("counter1", 10, 0)
		fmt.Println(pipe.Incr("counter1").Err())
		incr := redisClient.GetRange("counter1", 0, -1)
		fmt.Println(incr.Val(),"    ", incr.Err(), "   ", incr.String())
		return nil
	})
	fmt.Println(err)*/
	errCnt := setNX(key + KEY_COUNT)
	if errCnt != nil {
		return errCnt
	}

	errK := AddKey(key)
	if errK != nil {
		return errK
	}

	errP := LPush(key, mesg)

	if errP != nil {
		return errP
	}
	return nil
}

func updateMessage() {
	for {
		time.Sleep(10 * time.Second)
		keys, err := ZRange(CHAT_KEY)
		if err != nil {
			fmt.Println("Error getting keys ...........", err)
		}

		for i := 0; i < len(keys); i++ {
			msgs, merr := LRange(keys[i])
			if merr != nil {
				fmt.Println("Error getting List value ...........", merr)
			}
			//monMsgAr := make([]dao.ChatMessage, 0)
			for j := 0; j < len(msgs); j++ {
				chatMesg := ChatMsg{}
				unmarhal_srrr := json.Unmarshal([]byte(msgs[j]), &chatMesg)
				if unmarhal_srrr != nil {
					fmt.Println("Error in marshaling ...........", unmarhal_srrr)
					//return errr
				}
				id, seq_err := dao.GetNextSequence(CHAT_MESSAGE)
				if seq_err != nil {
					id = ""
				}
				switch chatMesg.Msg {
				case MESSAGE_CHAT:
					mongoMsg := dao.ChatMessage{Id: id, ChatId: keys[i],
						FromUserId:  chatMesg.FromClient,
						ToUserIds:   []string{chatMesg.ToClient},
						MessageType: strconv.FormatInt(chatMesg.Msg, 10),
						Message:     chatMesg.D,
						TimeStamp:   chatMesg.TimeStamp}
					UpdateMongoData(mongoMsg)
				case MESSAGE_CHAT_RO_ISSUE:
					mongoMsg := dao.ChatMessage{Id: id, ChatId: keys[i],
						FromUserId:  chatMesg.FromClient,
						ToUserIds:   []string{chatMesg.ToClient},
						MessageType: strconv.FormatInt(chatMesg.Msg, 10),
						Message:     chatMesg.D, ROId: chatMesg.ROId,
						TimeStamp: chatMesg.TimeStamp}
					UpdateMongoData(mongoMsg)
				}

				/*if minsert_error := dao.BulkInsert(mongoMsg); minsert_error != nil {
					fmt.Println("Error in Mongo insert ...........", minsert_error)
					//return err1
				}*/
				/*monMsgAr = append(monMsgAr, mongoMsg)
				fmt.Println("mongoMsg : ",mongoMsg)*/
			}
			delerr := DelKey(keys[i])
			if delerr != nil {
				fmt.Println("Key Not deleted -----> ", delerr)
			}

			/*fmt.Println("monMsAr : ",monMsgAr)
			fmt.Println("len monMsAr : ",len(monMsgAr))

			var interfaceSlice []interface{} = make([]interface{}, len(monMsgAr))
			for k, d := range monMsgAr {
				interfaceSlice[k] = d
			}
			dataObj := MongoData{MongoArr:monMsgAr}
			//fmt.Println("Before Updated -------> ", dataObj.MongoArr)
			if len(monMsgAr) != 0 {
				if err1 := dao.BulkInsert(dataObj); err1 != nil {
					fmt.Println("Error in Mongo insert ...........", err1)
				}else {
					//fmt.Println("Updated message -------> ", interfaceSlice)
				}
			}*/
		}
	}

}

func UpdateMongoData(chatMessage dao.ChatMessage) {
	if minsert_error := dao.Insert(chatMessage); minsert_error != nil {
		fmt.Println("Error in Mongo insert ...........", minsert_error)
		//return err1
	}
}
