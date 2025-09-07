package app

import (
    "crypto/tls"
    mgo "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "log"
    "net"
    "time"
)

const (
    // content of rest api
    COLLECTION = "content"
    // status collection for readiness probe
    STATUS = "status"
)

// struct for api content collection
type mgoApi struct {
    ID   bson.ObjectId `bson:"_id" json:"id"`
    Name string        `bson:"name" json:"name"`
}

// struct for status collection
type mgoStatus struct {
    ID     bson.ObjectId `bson:"_id" json:"id"`
    Status string        `bson:"status" json:"status"`
}

// struct for data access object
type contentsDAO struct {
    Servers  []string
    Database string
    Username string
    Password string
}

// mongodb connect function to establish connection to atlas
func (m *contentsDAO) Connect() {
    // enable tls
    tlsConfig := &tls.Config{}
    // set dail options
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
    // initiate mongodb atlas connection
    session, err := mgo.DialWithInfo(dialInfo)
    if err != nil {
        log.Fatal(err)
    }
    // set database name
    db = session.DB(m.Database)
}

// get all content function
func (m *contentsDAO) FindAll() ([]mgoApi, error) {
    var content []mgoApi
    err := db.C(COLLECTION).Find(bson.M{}).All(&content)
    return content, err
}

// find single content function
func (m *contentsDAO) FindById(id string) (mgoApi, error) {
    var content mgoApi
    err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&content)
    return content, err
}

// insert content function
func (m *contentsDAO) Insert(content mgoApi) error {
    err := db.C(COLLECTION).Insert(&content)
    return err
}

// update content function
func (m *contentsDAO) Update(content mgoApi) error {
    err := db.C(COLLECTION).UpdateId(content.ID, &content)
    return err
}

// delete content function
func (m *contentsDAO) Delete(content mgoApi) error {
    err := db.C(COLLECTION).Remove(&content)
    return err
}

// readiness probe to check database connection and get status content
func (m *contentsDAO) Readyz() ([]mgoStatus, error) {
    var status []mgoStatus
    err := db.C(STATUS).Find(bson.M{}).All(&status)
    return status, err
}

