package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

var (
	secretKey = []byte("secret-key")

//		progressData = `{
//		"appId": "0000010",
//		"active": "Y",
//		"done": [
//		  {
//			"lesson": "1",
//			"page": "12",
//			"result": "100%"
//		  },
//		  {
//			"lesson": "2",
//			"page": "12",
//			"result": "100%"
//		  },
//		  {
//			"lesson": "3",
//			"page": "",
//			"result": "0%"
//		  }
//		]
//	  }`
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

var progress Progress

func main() {

	// http.HandleFunc("/api/getToken", getToken)

	// http.HandleFunc("/api/recordProgress", recordProgress)

	// http.HandleFunc("/api/getProgress", getProgress)

	// log.Fatal(http.ListenAndServe(":8888", nil))

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/Token", getToken).Methods("GET")
	myRouter.HandleFunc("/Progress/{appId}", getProgress).Methods("GET")
	myRouter.HandleFunc("/Progress/{appId}", updProgress).Methods("PUT")
	// myRouter.HandleFunc("/article", createNewArticle).Methods("POST")
	// myRouter.HandleFunc("/article/{id}", returnSingleArticle)
	// myRouter.HandleFunc("/article/{id}", deleteArticle).Methods("DELETE")
	// myRouter.HandleFunc("/article/{id}", updateArticle).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8889", myRouter))

}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the InglesGuru API")
	fmt.Println("Endpoint Hit: InglesGuru API")
}

func getToken(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	//fmt.Printf("The request body is %v\n", r.Body)

	// var u User
	// json.NewDecoder(r.Body).Decode(&u)
	// fmt.Printf("The user request value %v", u)

	var a AppId
	json.NewDecoder(r.Body).Decode(&a)
	// fmt.Printf("The user request value %v", a)

	if a.AppId == "0000010" && a.Password == "123456" {
		tokenString, err := CreateToken(a.AppId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Application ID not found")
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, tokenString)
		return
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid credentials")
	}

	// if u.Username == "Nana" && u.Password == "123456" {
	// 	tokenString, err := CreateToken(u.Username)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		fmt.Fprint(w, "No User found")
	// 	}
	// 	w.WriteHeader(http.StatusOK)
	// 	fmt.Fprint(w, tokenString)
	// 	return
	// } else {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	fmt.Fprint(w, "Invalid credentials")
	// }

}

func getProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return
	}
	tokenString = tokenString[len("Bearer "):]

	err := verifyToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid token")
		return
	}

	// var p Progress
	// json.NewDecoder(r.Body).Decode(&progress)
	// fmt.Printf("The user request value %v", p)
	///////////////////////////////////

	vars := mux.Vars(r)
	appId := vars["appId"]

	if appId != "0000010" {
		fmt.Fprint(w, "Invalid App ID")
		return
	}

	// jData, err := json.Marshal(progressData)
	// if err != nil {
	// 	// handle error
	// }

	json.NewEncoder(w).Encode(progress)

	// w.Write(jData)

}

func updProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return
	}
	tokenString = tokenString[len("Bearer "):]

	err := verifyToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid token")
		return
	}

	vars := mux.Vars(r)
	appId := vars["appId"]

	if appId != "0000010" {
		fmt.Fprint(w, "Invalid App ID")
		return
	}

	// var p Progress
	json.NewDecoder(r.Body).Decode(&progress)
	// fmt.Printf("The user request value %v", p)

	//	progress = r.Body.Close().Error()

	fmt.Fprint(w, "Progress Recorded")

}

func CreateToken(appId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"appid": appId,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

func verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("Invalid token")
	}
	return nil
}
