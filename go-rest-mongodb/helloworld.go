package main

import (
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"os"
	"crypto/subtle"
)

const (
	ADMIN_USER     = "admin"
	ADMIN_PASSWORD = "password"
)

var server = os.Getenv("SERVER")
var database = os.Getenv("DATABASE")

func BasicAuth(handler http.HandlerFunc, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Content, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(Content),
			[]byte(ADMIN_USER)) != 1 || subtle.ConstantTimeCompare([]byte(pass),
			[]byte(ADMIN_PASSWORD)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("You are Unauthorized to access the application.\n"))
			return
		}
		handler(w, r)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld: received a request")

	response := os.Getenv("RESPONSE")

	if response == "" {
		response = "Hello, World!"
	}

	fmt.Fprintf(w, response+"\n"+os.Getenv("HOSTNAME"))
}

type Content struct {
	ID bson.ObjectId `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
}

type ContentsDAO struct {
	Server   string
	Database string
}

var dao = ContentsDAO{}
var db *mgo.Database

const (
	COLLECTION = "Contents"
)

// Establish a connection to database
func (m *ContentsDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

// Find list of Contents
func (m *ContentsDAO) FindAll() ([]Content, error) {
	var Contents []Content
	err := db.C(COLLECTION).Find(bson.M{}).All(&Contents)
	return Contents, err
}

// Find a Content by its id
func (m *ContentsDAO) FindById(id string) (Content, error) {
	var Content Content
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&Content)
	return Content, err
}

// Insert a Content into database
func (m *ContentsDAO) Insert(Content Content) error {
	err := db.C(COLLECTION).Insert(&Content)
	return err
}

// Delete an existing Content
func (m *ContentsDAO) Delete(Content Content) error {
	err := db.C(COLLECTION).Remove(&Content)
	return err
}

// Update an existing Content
func (m *ContentsDAO) Update(Content Content) error {
	err := db.C(COLLECTION).UpdateId(Content.ID, &Content)
	return err
}

// GET list of Contents
func AllContentsEndPoint(w http.ResponseWriter, r *http.Request) {
	Contents, err := dao.FindAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, Contents)
}

// GET a Contents by its ID
func FindContentEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	Content, err := dao.FindById(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content ID")
		log.Print(err)
		return
	}
	respondWithJson(w, http.StatusOK, Content)
}

// POST a new Content
func CreateContentEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var Content Content
	if err := json.NewDecoder(r.Body).Decode(&Content); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		log.Print(err)
		return
	}
	Content.ID = bson.NewObjectId()
	if err := dao.Insert(Content); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusCreated, Content)
}

// PUT update an existing Content
func UpdateContentEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	params := mux.Vars(r)
	var Content Content
	Content.ID = bson.ObjectIdHex(params["id"])
	if err := json.NewDecoder(r.Body).Decode(&Content); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.Update(Content); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}

// DELETE an existing Content
func DeleteContentEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var Content Content
	if err := json.NewDecoder(r.Body).Decode(&Content); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.Delete(Content); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func init() {

	if server == "" {
		server = "mongodb"
	}

	if database == "" {
		database = "contents_db"
	}

	dao.Server = server
	dao.Database = database
	dao.Connect()
}

func main() {
	log.Print("helloworld: is starting...")
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", handler)
	r.HandleFunc("/api/v1/content", BasicAuth(AllContentsEndPoint, "Please enter your username and password")).Methods("GET")
	r.HandleFunc("/api/v1/content", BasicAuth(CreateContentEndPoint, "Please enter your username and password")).Methods("POST")
	r.HandleFunc("/api/v1/content", BasicAuth(DeleteContentEndPoint, "Please enter your username and password")).Methods("DELETE")
	r.HandleFunc("/api/v1/content/{id}", BasicAuth(UpdateContentEndPoint, "Please enter your username and password")).Methods("PUT")
	r.HandleFunc("/api/v1/content/{id}", BasicAuth(FindContentEndpoint, "Please enter your username and password")).Methods("GET")
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Printf("helloworld: listening on port %s", port)
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("error starting http server : ", err)
	}
}
