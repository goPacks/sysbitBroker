package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sysbitBroker/auth"
	"sysbitBroker/data"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
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
	conn *pgx.Conn
	err  error
)

const (
	connStr = "postgres://postgres:mysecretpassword@143.198.198.51:5432/inglesapp?sslmode=disable"
)

func main() {

	conn, err = pgx.Connect(context.Background(), connStr)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer conn.Close(context.Background())

	// Create a Router without the Token Authenitcation
	router := mux.NewRouter()
	router.HandleFunc("/AppToken", getAppToken).Methods("POST")
	router.HandleFunc("/AdminToken", getAdminToken).Methods("POST")
	router.HandleFunc("/chkApi", chkApi).Methods("GET")

	// Create Admin subRouter with Token Authentication
	adminRouter := router.PathPrefix("/").Subrouter()
	adminRouter.Use(chkAdminToken)
	adminRouter.HandleFunc("/getQuiz/{quizId}", getQuiz).Methods("GET")
	adminRouter.HandleFunc("/updQuiz/{quizId}", updQuiz).Methods("PUT")
	adminRouter.HandleFunc("/getLesson/{lessonId}", getLesson).Methods("GET")
	adminRouter.HandleFunc("/updLesson/{lessonId}", updLesson).Methods("PUT")

	//defining authenticated route
	appRouter := router.PathPrefix("/").Subrouter()
	appRouter.Use(chkAppToken)

	// Register the routes on the main router with the auth chkToken
	appRouter.HandleFunc("/regApp/{appId}", regApp).Methods("POST")
	appRouter.HandleFunc("/getAppInfo/{appId}", getAppInfo).Methods("GET")
	appRouter.HandleFunc("/updAppInfo/{appId}", updAppInfo).Methods("PUT")

	fmt.Println("Server Listening on port 8899")
	log.Fatal(http.ListenAndServe(":8899", router))

}

func chkAppToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := auth.ChkAppToken(w, r)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		h.ServeHTTP(w, r)
	})
}

func chkAdminToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := auth.ChkAdminToken(w, r)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		h.ServeHTTP(w, r)
	})
}

func chkApi(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the InglesGuru API")
	fmt.Println("Endpoint Hit: InglesGuru API")
}

func getAppToken(w http.ResponseWriter, r *http.Request) {
	auth.GetAppToken(w, r, conn)
}

func getAdminToken(w http.ResponseWriter, r *http.Request) {
	auth.GetAdminToken(w, r, conn)
}

func getConv(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	lessonId := vars["alessonId"]

	data.GetConv(w, r, conn, lessonId)
}

func getQuiz(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	quizId := vars["quizId"]

	data.GetQuiz(w, r, conn, quizId)
}

func getLesson(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	lessonId := vars["lessonId"]

	data.GetLesson(w, r, conn, lessonId)
}

func updQuiz(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	quizId := vars["quizId"]

	data.UpdQuiz(w, r, conn, quizId)
}

func updLesson(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	lessonId := vars["lessonId"]

	data.UpdLesson(w, r, conn, lessonId)
}

func getAppInfo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	appId := vars["appId"]

	fmt.Println("Here")

	data.GetAppInfo(w, r, conn, appId)
}

func regApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	appId := vars["appId"]

	data.RegisterApp(w, r, conn, appId)
}

func updAppInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	appId := vars["appId"]

	data.UpdAppInfo(w, r, conn, appId)
}
