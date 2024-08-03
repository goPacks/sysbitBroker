package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sysbitBroker/auth"
	"sysbitBroker/data"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
)

var (
	dsn = ""
	cnt int64
)

type AppId struct {
	AppId    string `json:"appid"`
	Password string `json:"password"`
}

type Lesson struct {
	Lesson string `json:"Lesson"`
	Page   string `json:"Page"`
	Result string `json:"Result"`
}

type Progress struct {
	AppId  string   `json:"appid"`
	Active string   `json:"active"`
	Done   []Lesson `json:"done"`
}

type OKReplyProgress struct {
	Status string
	Data   Progress
}

type OKReply struct {
	Status  string
	Message string
}

type NOKReply struct {
	Status string
	Errors string
}

var (
	progress Progress
	conn     *pgx.Conn
)

// const (
// 	host     = "localhost"
// 	port     = 5400
// 	user     = "postgres"
// 	password = "mysecretpassword"
// 	dbname   = "db_1"
// )

//connStr := "postgres://postgres:mysecretpassword@localhost/db_1?sslmode=disable"

func main() {

	var err error

	conn, err = pgx.Connect(context.Background(), "postgres://postgres:mysecretpassword@localhost/db_1?sslmode=disable")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer conn.Close(context.Background())

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", HomePage)
	myRouter.HandleFunc("/Token", GetToken).Methods("GET")
	myRouter.HandleFunc("/InfoApp/{appId}", InfoApp).Methods("GET")
	myRouter.HandleFunc("/UpdProgress/{appId}", UpdProgress).Methods("PUT")
	myRouter.HandleFunc("/AddApp/{appId}", AddAppId).Methods("POST")

	log.Fatal(http.ListenAndServe(":8899", myRouter))

}

func HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the InglesGuru API")
	fmt.Println("Endpoint Hit: InglesGuru API")
}

func GetToken(w http.ResponseWriter, r *http.Request) {
	auth.GetToken(w, r, conn)
}

func InfoApp(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if err := auth.CheckToken(w, r); err != nil {
		ReturnError(w, err.Error())
		return
	}

	vars := mux.Vars(r)
	appId := vars["appId"]

	data.InfoApp(w, r, conn, appId)
}

func AddAppId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := auth.CheckToken(w, r); err != nil {
		ReturnError(w, err.Error())
		return
	}

	vars := mux.Vars(r)
	appId := vars["appId"]

	data.AddApp(w, r, conn, appId)
}

func UpdProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := auth.CheckToken(w, r); err != nil {
		ReturnError(w, err.Error())
		return
	}

	vars := mux.Vars(r)
	appId := vars["appId"]

	data.UpdProgress(w, r, conn, appId)
}

func ReturnError(w http.ResponseWriter, strErr string) {
	var nokReply NOKReply
	nokReply.Status = "NOK"
	nokReply.Errors = strErr
	json.NewEncoder(w).Encode(nokReply)
}
