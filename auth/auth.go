package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
)

const adminTokenSecret = "Hello, 世界"

var (
	secretKey = []byte("secret-key")
)

type AppId struct {
	AppId string `json:"appId"`
	Pin   string `json:"pin"`
}

type Login struct {
	UserId string `json:"userId"`
	Pswd   string `json:"pswd"`
}

type NOKReply struct {
	Status string
	Errors string
}

type OKReply struct {
	Status string
	Token  string
}

func GetAppToken(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {

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

func GetAdminToken(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {

	w.Header().Set("Content-Type", "application/json")

	var a Login

	json.NewDecoder(r.Body).Decode(&a)

	name := ""
	email := ""
	roleGroupId := ""

	if err := conn.QueryRow(context.Background(), "select name, roleGroupId, email from login where loginId = $1 and pswd = $2", a.UserId, a.Pswd).Scan(&name, &roleGroupId, &email); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	tokenString, err := createToken(adminTokenSecret)
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

func ChkAppToken(w http.ResponseWriter, r *http.Request) error {
	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		//w.WriteHeader(http.StatusUnauthorized)
		return errors.New("Missing Authorization Header")
	}

	tokenString = tokenString[len("Bearer "):]

	vars := mux.Vars(r)
	appId := vars["appId"]

	err := verifyToken(tokenString, appId)
	if err != nil {
		//w.WriteHeader(http.StatusUnauthorized)
		//return errors.New("Invalid Token")
		return err

	}

	return nil
}

func ChkAdminToken(w http.ResponseWriter, r *http.Request) error {
	tokenString := r.Header.Get("Authorization")

	if tokenString == "" {
		//w.WriteHeader(http.StatusUnauthorized)
		return errors.New("Missing Authorization Header")
	}

	tokenString = tokenString[len("Bearer "):]

	err := verifyToken(tokenString, adminTokenSecret)
	if err != nil {
		//w.WriteHeader(http.StatusUnauthorized)
		//return errors.New("Invalid Token")
		return err

	}

	return nil
}

func createToken(appId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"appId": appId,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

func verifyToken(tokenString string, appId string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("Invalid Token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("Unable to extract claims")
	}

	if claims["appId"].(string) != appId {
		return fmt.Errorf("Token does not match appID")
	}

	return nil
}
