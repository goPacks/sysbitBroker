package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sysbitBroker/auth"
	"sysbitBroker/data"

	"encoding/json"
	"errors"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
)

var (
	secretKey = []byte("secret-key")
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
	router.HandleFunc("/regApp", regApp).Methods("POST")

	tokenRouter := router.PathPrefix("/").Subrouter()
	tokenRouter.Use(chkToken)
	tokenRouter.HandleFunc("/getHeaders/{modCode}", getHeaders).Methods("GET")
	tokenRouter.HandleFunc("/getQuiz/{quizCode}", getQuiz).Methods("GET")
	tokenRouter.HandleFunc("/updQuiz/{quizCode}", updQuiz).Methods("PUT")
	tokenRouter.HandleFunc("/getLesson/{lessonCode}", getLesson).Methods("GET")
	tokenRouter.HandleFunc("/updLesson/{lessonCode}", updLesson).Methods("PUT")
	tokenRouter.HandleFunc("/regApp", regApp).Methods("POST")
	tokenRouter.HandleFunc("/getAppProgress", getAppProgress).Methods("GET")
	tokenRouter.HandleFunc("/updAppProgress", updAppProgress).Methods("PUT")

	// Create Admin subRouter with Token Authentication
	// adminRouter := router.PathPrefix("/").Subrouter()
	// adminRouter.Use(chkAdminToken)
	// adminRouter.HandleFunc("/getQuiz/{quizId}", getQuiz).Methods("GET")
	// adminRouter.HandleFunc("/updQuiz/{quizId}", updQuiz).Methods("PUT")
	// adminRouter.HandleFunc("/getLesson/{lessonId}", getLesson).Methods("GET")
	// adminRouter.HandleFunc("/updLesson/{lessonId}", updLesson).Methods("PUT")
	// //	adminRouter.HandleFunc("/AdminLessonHeaders/{moduleCode}", getLessonHeaders).Methods("GET")

	// //defining authenticated route
	// appRouter := router.PathPrefix("/").Subrouter()
	// appRouter.Use(chkAppToken)
	// // Register the routes on the main router with the auth chkToken
	// appRouter.HandleFunc("/regApp", regApp).Methods("POST")
	// appRouter.HandleFunc("/getAppInfo", getAppInfo).Methods("GET")
	// appRouter.HandleFunc("/updAppInfo", updAppInfo).Methods("PUT")
	// //	appRouter.HandleFunc("/AppLessonHeaders/{moduleCode}", getLessonHeaders).Methods("GET")

	fmt.Println("Server Listening on port 8899")
	log.Fatal(http.ListenAndServe(":8899", router))

}

// func chkAppToken(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		err := auth.ChkAppToken(h, w, r)

// 		if err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}

// 		h.ServeHTTP(w, r)
// 	})
// }

// func chkAdminToken(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		err := auth.ChkAdminToken(w, r)

// 		if err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}

// 		h.ServeHTTP(w, r)
// 	})
// }

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
	quizCode := vars["quizCode"]

	data.GetQuiz(w, r, conn, quizCode)
}

func getLesson(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	lessonCode := vars["lessonCode"]

	data.GetLesson(w, r, conn, lessonCode)
}

func getHeaders(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	modCode := vars["modCode"]

	data.GetHeaders(w, r, conn, modCode)
}

func updQuiz(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	quizCode := vars["quizCode"]

	data.UpdQuiz(w, r, conn, quizCode)
}

func updLesson(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	lessonCode := vars["lessonCode"]

	data.UpdLesson(w, r, conn, lessonCode)
}

func getAppProgress(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// vars := mux.Vars(r)
	// appId := vars["appId"]

	data.GetAppProgress(w, r, conn)
}

func regApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data.RegisterApp(w, r, conn)
}

func updAppProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data.UpdAppProgress(w, r, conn)
}

// func chkAppToken(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		tokenString := r.Header.Get("Authorization")
// 		if tokenString == "" {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			return secretKey, nil
// 		})

// 		if err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		if !token.Valid {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		appId, ok := claims["appId"].(string)
// 		if !ok {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), "appId", appId)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

// func chkAdminToken(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		tokenString := r.Header.Get("Authorization")
// 		if tokenString == "" {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			return secretKey, nil
// 		})

// 		if err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		if !token.Valid {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		// name, ok1 := claims["name"].(string)
// 		// roleCode, ok2 := claims["name"].(string)
// 		email, ok3 := claims["email"].(string)

// 		if !ok1 || !ok2 || !ok3 {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		// ctx := context.WithValue(r.Context(), "name", name)
// 		// ctx = context.WithValue(r.Context(), "roleCode", roleCode)
// 		ctx := context.WithValue(r.Context(), "email", email)

// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func chkToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			errMsg(w, errors.New("no token"), http.StatusUnauthorized)
			return
		}

		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil {
			errMsg(w, err, http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			errMsg(w, errors.New("token not valid"), http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			errMsg(w, errors.New("claims not valid"), http.StatusUnauthorized)
			return
		}

		tokenType := claims["tokenType"].(string)

		if tokenType == "admin" {
			name, ok1 := claims["name"].(string)
			roleCode, ok2 := claims["roleCode"].(string)
			email, ok3 := claims["email"].(string)

			if !ok1 {
				errMsg(w, errors.New("name in claims not valid"), http.StatusUnauthorized)
				return
			}

			if !ok2 {
				errMsg(w, errors.New("rolecode in claims not valid"), http.StatusUnauthorized)
				return
			}

			if !ok3 {
				errMsg(w, errors.New("email in claims not valid"), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "name", name)
			ctx = context.WithValue(ctx, "roleCode", roleCode)
			ctx = context.WithValue(ctx, "email", email)

			next.ServeHTTP(w, r.WithContext(ctx))

		} else {

			appId, ok := claims["appId"].(string)

			if !ok {
				errMsg(w, errors.New("appId in claims not valid"), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "appId", appId)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

	})
}

func errMsg(w http.ResponseWriter, err error, status int) {
	var nokReply NOKReply
	nokReply.Status = "NOK"
	nokReply.Errors = err.Error()
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(nokReply)

}
