package main

import (
	"encoding/json"
	"log"
	"net/http"
	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"os"
	"crypto/subtle"
)

// Represents database server and credentials
// type Config struct {
// 	Server   string
// 	Database string
// }
//
// var config = Config{"localhost", "users_db"}

const (
	ADMIN_USER     = "admin"
	ADMIN_PASSWORD = "password"
)

func BasicAuth(handler http.HandlerFunc, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user),
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

var server = os.Getenv("SERVER")
var database = os.Getenv("DATABASE")

type User struct {
	ID bson.ObjectId `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	Age  int `bson:"age" json:"age"`
	Email string `bson:"email" json:"email"`
}

type UsersDAO struct {
	Server   string
	Database string
}

var dao = UsersDAO{}
var db *mgo.Database

const (
	COLLECTION = "users"
)

// Establish a connection to database
func (m *UsersDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

// Find list of users
func (m *UsersDAO) FindAll() ([]User, error) {
	var users []User
	err := db.C(COLLECTION).Find(bson.M{}).All(&users)
	return users, err
}

// Find a user by its id
func (m *UsersDAO) FindById(id string) (User, error) {
	var user User
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&user)
	return user, err
}

// Insert a user into database
func (m *UsersDAO) Insert(user User) error {
	err := db.C(COLLECTION).Insert(&user)
	return err
}

// Delete an existing user
func (m *UsersDAO) Delete(user User) error {
	err := db.C(COLLECTION).Remove(&user)
	return err
}

// Update an existing user
func (m *UsersDAO) Update(user User) error {
	err := db.C(COLLECTION).UpdateId(user.ID, &user)
	return err
}

// GET list of users
func AllUsersEndPoint(w http.ResponseWriter, r *http.Request) {
	users, err := dao.FindAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, users)
}

// GET a users by its ID
func FindUserEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user, err := dao.FindById(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		log.Print(err)
		return
	}
	respondWithJson(w, http.StatusOK, user)
}

// POST a new user
func CreateUserEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		log.Print(err)
		return
	}
	user.ID = bson.NewObjectId()
	if err := dao.Insert(user); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusCreated, user)
}

// PUT update an existing user
func UpdateUserEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	params := mux.Vars(r)
	var user User
	user.ID = bson.ObjectIdHex(params["id"])
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.Update(user); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}

// DELETE an existing user
func DeleteUserEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.Delete(user); err != nil {
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

// Parse the configuration file 'config.toml', and establish a connection to DB
func init() {
	dao.Server = server
	dao.Database = database
	dao.Connect()
}

// Define HTTP request routes
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/users", BasicAuth(AllUsersEndPoint, "Please enter your username and password")).Methods("GET")
	r.HandleFunc("/users", BasicAuth(CreateUserEndPoint, "Please enter your username and password")).Methods("POST")
	r.HandleFunc("/users/{id}", BasicAuth(UpdateUserEndPoint, "Please enter your username and password")).Methods("PUT")
	r.HandleFunc("/users", BasicAuth(DeleteUserEndPoint, "Please enter your username and password")).Methods("DELETE")
	r.HandleFunc("/users/{id}", BasicAuth(FindUserEndpoint, "Please enter your username and password")).Methods("GET")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
