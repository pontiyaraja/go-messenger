package utills

import (
	"encoding/json"
	"net/http"
	"reflect"
	"bitbucket.org/tekion/tekionbaas/log"
	"strings"
)

func AuthenticateUser(clientId string, accessToken string) string {
	req, _ := http.NewRequest(http.MethodGet, AUTHENTICATE_URL, nil)
	req.Header.Set(TEKION_API_TOKEN, accessToken)
	client := &http.Client{}
	resp, _ := client.Do(req)
	var response interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	i := reflect.ValueOf(response).Interface()
	log.Info(i)
	resMap := i.(map[string]interface{})
	metaMap := resMap["meta"].(map[string]interface{})
	code := metaMap["code"].(float64)
	msg := metaMap["msg"]
	if code == 200 {
		dataMap := resMap["data"].(map[string]interface{})
		if dataMap["id"] == clientId || strings.Compare(dataMap["email"].(string), clientId) == 0{
			return "success"
		}
	}
	return msg.(string)
}
