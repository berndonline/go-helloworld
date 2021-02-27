package main

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

const (
	COLLECTION = "content"
)

type contentsDAO struct {
	Server   string
	Database string
}

func (m *contentsDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

func (m *contentsDAO) FindAll() ([]mgoApi, error) {
	var content []mgoApi
	err := db.C(COLLECTION).Find(bson.M{}).All(&content)
	return content, err
}

func (m *contentsDAO) FindById(id string) (mgoApi, error) {
	var content mgoApi
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&content)
	return content, err
}

func (m *contentsDAO) Insert(content mgoApi) error {
	err := db.C(COLLECTION).Insert(&content)
	return err
}

func (m *contentsDAO) Delete(content mgoApi) error {
	err := db.C(COLLECTION).Remove(&content)
	return err
}

func (m *contentsDAO) Update(content mgoApi) error {
	err := db.C(COLLECTION).UpdateId(content.ID, &content)
	return err
}
