package dao

import (
	"fmt"
	"go/ast"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strings"
	"sync"
)

//var DB *mgo.Database
//var UserC *mgo.Collection
var todoC *mgo.Collection
var vinMasterC *mgo.Collection
var customerVinC *mgo.Collection

//var SequenceC *mgo.Collection
var db *DB

type CollectionListMap struct {
	m map[string]*mgo.Collection
	l *sync.RWMutex
}

func (c *CollectionListMap) Set(key string, value *mgo.Collection) {

	c.l.Lock()
	defer c.l.Unlock()
	c.m[key] = value

}

func (c *CollectionListMap) Get(key string) (*mgo.Collection, error) {

	c.l.RLock()
	defer c.l.RUnlock()
	item, ok := c.m[key]
	if !ok {
		return &mgo.Collection{}, fmt.Errorf("'%s' is not present in CollectionList", key)
	}
	return item, nil
}

var (
	CollectionList *CollectionListMap
	once           sync.Once
)

//For singleton behaviour of CollectionList
func NewCollectionList() *CollectionListMap {

	once.Do(func() {
		CollectionList = &CollectionListMap{
			l: new(sync.RWMutex),
			m: make(map[string]*mgo.Collection),
		}
	})
	return CollectionList
}

func init() {
	ConnectMongo()
	createCollection(Sequence{}, []string{"_id"})
	createCollection(ChatMessage{}, []string{"_id"})
}

func ConnectMongo() {
	dataBase, err := Connect()

	//Handle panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Detected panic 1")
			var ok bool
			err, ok := r.(error)
			if !ok {
				fmt.Printf("pkg:  %v,  error: %s", r, err)
			}
		}
	}()

	db = dataBase
	if err != nil {
		panic("Mongo is not connected")
		return
	}

	//Todo Refactor it in better way
	createSessionForCollection(db)

}

//Instead of using global variables, maintain session for each collection in CollectionList
func createSessionForCollection(db *DB) {
	CollectionList = NewCollectionList()

	ChatMessageC := db.Db.C("ChatMessage")
	SequenceC := db.Db.C("Sequence")

	CollectionList.Set("ChatMessage", ChatMessageC)
	CollectionList.Set("Sequence", SequenceC)
}

func createCollection(obj interface{}, index []string) {

	collectionName := reflect.TypeOf(obj).Name()

	//Handle panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Detected panic 2")
			var ok bool
			err, ok := r.(error)
			if !ok {
				fmt.Printf("pkg:  %v,  error: %s", r, err)
			}
		}
	}()

	c := db.Db.C(collectionName)

	idx := mgo.Index{
		Key:        index,
		Unique:     false,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(idx)
	if err != nil {
		panic(err)
	}
	CollectionList.Set(collectionName, c)
}

func getStructFields(value interface{}) ([]string, error) {
	fields := []string{}

	reflectType := reflect.ValueOf(value).Type()

	if reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	if reflectType.Kind() != reflect.Struct {
		strMsg := "Must be of struct type"
		panic(strMsg)
	}

	for i := 0; i < reflectType.NumField(); i++ {
		if fieldStruct := reflectType.Field(i); ast.IsExported(fieldStruct.Name) {

			indirectType := fieldStruct.Type
			for indirectType.Kind() == reflect.Ptr {
				indirectType = indirectType.Elem()
			}

			//if (indirectType.Kind() != reflect.Slice) {
			fmt.Println(fieldStruct.Name)
			tag := tagParsing(fieldStruct.Tag)
			fields = append(fields, tag)
			//}
		}
	}

	return fields, nil

}

func tagParsing(tag reflect.StructTag) string {

	bsonTag := ""

	for _, str := range []string{tag.Get("bson")} {
		subTags := strings.Split(str, ",")

		if len(subTags) > 0 {
			bsonTag = subTags[0]
		} else {
			err := "bson tag is missiing"
			panic(err)
		}
	}
	return bsonTag
}

func fieldSelector(q ...string) (r bson.M) {
	r = make(bson.M, len(q))
	for _, s := range q {
		r[s] = 1
	}
	return
}
