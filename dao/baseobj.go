package dao

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"reflect"
)

type Dao interface {
	insert()
	update()
	find()
	get()
}

type DB struct {
	Db    *mgo.Database
	Value interface{}
}

func Connect() (*DB, error) {
	var db DB

	//mgo.SetDebug(true)
	//var aLogger *log.Logger
	//aLogger = log.New(os.Stderr, "", log.LstdFlags)
	//mgo.SetLogger(aLogger)

	//session, err := mgo.Dial("127.0.0.1") // Get this value from configuration
	session, err := mgo.Dial("10.0.0.167") // Get this value from configuration
	if err != nil {
		return &db, err
	}

	d := session.DB(DATABASE_NAME)

	db = DB{Db: d, Value: nil}
	return &db, nil
}

func getFieldFromStruct(obj interface{}, field string) string {
	v := reflect.ValueOf(obj).Elem().FieldByName("id").String()
	return v
}

func GetNextId(obj interface{}) (string, error) {
	collectionName := reflect.TypeOf(obj).Name()
	seq, err := GetNextSequence(collectionName)
	if err != nil {
		return "", err
	}
	return seq, err
}

func Insert(obj interface{}) error {
	collectionName := reflect.TypeOf(obj).Name()
	//v := reflect.ValueOf(&obj).Elem().Elem().FieldByName("Id").String()
	/*fmt.Println(v)
	objFound, err := Find(obj,v)
	if err != nil {
		return err
	}
	if objFound {
		return types.ErrDuplicateRecord
	}*/
	collection, err1 := CollectionList.Get(collectionName)

	if err1 != nil {
		fmt.Println("Collection doesn't exists")
		return err1
	}

	err := collection.Insert(obj)
	return err
}

func Find(obj interface{}, id string) (bool, error) {

	if id == "" {
		return false, ErrInvalidObject
	}
	uid, err := GetId(obj, id)

	objFound := false
	if err != nil && err == mgo.ErrNotFound {
		objFound = false
		err = nil
	}

	if err == nil && uid == uid {
		objFound = true
	}

	return objFound, err
}

func Get(obj interface{}, id string) (interface{}, error) {
	typeOfT := reflect.TypeOf(obj)
	result := reflect.New(typeOfT)
	err := CollectionList.m[typeOfT.Name()].FindId(id).One(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func GetId(obj interface{}, id string) (string, error) {
	fmt.Println("Entering GetId function ")

	//Handle panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Detected panic")
			var ok bool
			err, ok := r.(error)
			if !ok {
				fmt.Printf("pkg:  %v,  error: %s", r, err)
			}
		}
	}()

	typeOfT := reflect.TypeOf(obj)
	result := reflect.New(typeOfT)
	err := CollectionList.m[typeOfT.Name()].FindId(id).One(&result)
	if err != nil {
		return "", err
	}

	if result.Kind() != reflect.Struct {
		panic("id must belong to struct")
	}
	v := result.FieldByName("id") // This function would Panic on error should handle gracefully -- @Mani
	return v.String(), nil
}

func Update(obj Dao) {
	obj.update()
}

type DaoBase struct {
	Id string `bson:"_id" json:"id"`
}

type User2 struct {
	DaoBase
	FName string `bson:"fname" json:"fname"`
}

func (u *User2) insert() {
	Insert(u)
}
