package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
)

var (
	secretKey = []byte("secret-key")
)

type AppId struct {
	AppId string `json:"appid"`
	Pin   string `json:"pin"`
}

type NOKReply struct {
	Status string
	Errors string
}

type OKReply struct {
	Status string
	Token  string
}

func GetToken(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {

	w.Header().Set("Content-Type", "application/json")

	var a AppId

	json.NewDecoder(r.Body).Decode(&a)

	tokenString, err := createToken(a.AppId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = "Error Creating Token"
		json.NewEncoder(w).Encode(nokReply)
		return

	}
	w.WriteHeader(http.StatusOK)
	var okReply OKReply
	okReply.Status = "OK"
	okReply.Token = tokenString
	json.NewEncoder(w).Encode(okReply)

}

// func checkPermission(h http.Handler) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		authCheck := true

// 		if authCheck {

// 			w.WriteError(w, 400, "error")
// 			return
// 		}

// 		h.ServeHttp(w, r)
// 	}
// }

func ChkToken(w http.ResponseWriter, r *http.Request) error {
	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		//w.WriteHeader(http.StatusUnauthorized)
		return errors.New("Missing Authorization Header")
	}

	tokenString = tokenString[len("Bearer "):]

	err := verifyToken(tokenString)
	if err != nil {
		//w.WriteHeader(http.StatusUnauthorized)
		return errors.New("Invalid Token")

	}

	return nil
}

func createToken(appId string) (string, error) {
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
