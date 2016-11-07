package dao

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strconv"
)

type Sequence struct {
	SequenceName string `bson:"sequenceName" json:"sequenceName"`
	Next         uint64 `bson:next" json:nex"`
}

func InsertSequence(sequenceName string) error {

	seqFound, err := FindSequence(sequenceName)
	if err != nil {
		return err
	}

	if seqFound {
		return ErrDuplicateRecord
	}
	s := Sequence{}
	s.SequenceName = sequenceName
	s.Next = 0

	collectionName := reflect.TypeOf(s).Name()
	SequenceC, err1 := CollectionList.Get(collectionName)

	if err1 != nil {
		return err1
	}
	err = SequenceC.Insert(s)
	return err

}

func GetNextSequence(sequenceName string) (string, error) {
	seqFound, err := FindSequence(sequenceName)
	if err != nil {
		return "0", err
	}
	if !seqFound {
		err = InsertSequence(sequenceName)
		if err != nil {
			return "0", err
		}
	}
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"next": 1}},
		ReturnNew: true,
	}

	seq := Sequence{}
	collectionName := reflect.TypeOf(seq).Name()
	sequence, err1 := CollectionList.Get(collectionName)

	if err1 != nil {
		return "0", err1
	}
	_, err = sequence.Find(bson.M{"sequenceName": sequenceName}).Apply(change, &seq)
	return strconv.FormatUint(seq.Next, 10), err
}

func FindSequence(sequenceName string) (bool, error) {

	if sequenceName == "" {
		return false, ErrInvalidObject
	}
	sName, err := GetSequence(sequenceName)

	objFound := false
	if err != nil && err == mgo.ErrNotFound {
		objFound = false
		err = nil
	}

	if err == nil && sequenceName == sName {
		objFound = true
	}

	return objFound, err
}

func GetSequence(sequenceName string) (string, error) {
	result := &Sequence{}

	collectionName := reflect.TypeOf(Sequence{}).Name()
	sequence, err1 := CollectionList.Get(collectionName)

	if err1 != nil {
		return "0", err1
	}

	err := sequence.Find(bson.M{"sequenceName": sequenceName}).One(&result)
	if err != nil {
		return "", err
	}

	return result.SequenceName, err
}
