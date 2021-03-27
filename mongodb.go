package main

import (
	"crypto/tls"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net"
	"time"
)

const (
	COLLECTION = "content"
)

type contentsDAO struct {
	Servers  []string
	Database string
	Username string
	Password string
}

func (m *contentsDAO) Connect() {
	tlsConfig := &tls.Config{}
	dialInfo := &mgo.DialInfo{
		Addrs:    m.Servers,
		Timeout:  60 * time.Second,
		Username: m.Username,
		Password: m.Password,
	}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}
	session, err := mgo.DialWithInfo(dialInfo)
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
