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
	//	conn, err = pgx.Connect(context.Background(), "postgres://postgres:sysbitDB@localhost/db_1?sslmode=disable")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer conn.Close(context.Background())

	router := mux.NewRouter()

	// Create a new router without the auth chkToken
	router.HandleFunc("/token", getToken).Methods("POST")
	router.HandleFunc("/chkApi", chkApi).Methods("GET")

	//defining authenticated route
	privateRouter := router.PathPrefix("/").Subrouter()
	privateRouter.Use(chkToken)

	// Register the routes on the main router with the auth chkToken
	privateRouter.HandleFunc("/regApp/{appId}", regApp).Methods("POST")
	privateRouter.HandleFunc("/getAppInfo/{appId}", getAppInfo).Methods("GET")
	privateRouter.HandleFunc("/updAppInfo/{appId}", updAppInfo).Methods("PUT")

	fmt.Println("Server Listening on port 8899")
	log.Fatal(http.ListenAndServe(":8899", router))

}

func chkToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// tokenString := r.Header.Get("Authorization")
		// if tokenString == "" {
		//     w.WriteHeader(http.StatusUnauthorized)
		//     return
		// }

		// tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		// token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//     if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		//         return nil, fmt.Errorf("unexpected signing method")
		//     }
		//     return []byte("secret"), nil
		// })

		// if err != nil {
		//     w.WriteHeader(http.StatusUnauthorized)
		//     return
		// }

		// if !token.Valid {
		//     w.WriteHeader(http.StatusUnauthorized)
		//     return
		// }

		// claims, ok := token.Claims.(jwt.MapClaims)
		// if !ok {
		//     w.WriteHeader(http.StatusUnauthorized)
		//     return
		// }

		// userID, ok := claims["user_id"].(string)
		// if !ok {
		//     w.WriteHeader(http.StatusUnauthorized)
		//     return
		// }

		// ctx := context.WithValue(r.Context(), "user_id", userID)

		//authCheck := true
		err := auth.ChkToken(w, r)

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

func getToken(w http.ResponseWriter, r *http.Request) {
	auth.GetToken(w, r, conn)
}

func getAppInfo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	appId := vars["appId"]

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

// func ReturnError(w http.ResponseWriter, strErr string) {
// 	var nokReply NOKReply
// 	nokReply.Status = "NOK"
// 	nokReply.Errors = strErr
// 	json.NewEncoder(w).Encode(nokReply)
// }
