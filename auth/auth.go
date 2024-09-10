package auth

import (
	"context"
	"encoding/json"

	"net/http"
	"sysbitBroker/errHandler"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
)

var (
	secretKey = []byte("keysysbitkey")
)

type AppId struct {
	AppId string `json:"appId"`
	Email string `json:"email"`
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

	pmtLevel := 0
	active := false
	nativeLingo := ""
	deviceOs := ""

	if err := conn.QueryRow(context.Background(), "select pmtLevel, active, nativeLingo, deviceOs from app where appId = $1 and email = $2", a.AppId, a.Email).Scan(&pmtLevel, &active, &nativeLingo, &deviceOs); err != nil {
		errHandler.ErrMsg(w, err, 404)
		return
	}

	claims := jwt.MapClaims{
		"tokenType":    "app",
		"appId":        a.AppId,
		"email":        a.Email,
		"pmtLevel":     pmtLevel,
		"active":       active,
		"nartiveLingo": nativeLingo,
		"deviceOs":     deviceOs,
		"exp":          time.Now().Add(time.Hour * 24).Unix(),
	}

	tokenString, err := createToken(claims)
	if err != nil {
		errHandler.ErrMsg(w, err, 404)
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
	roleCode := ""

	if err := conn.QueryRow(context.Background(), "select name, roleCode, email from login where loginCode = $1 and pswd = $2", a.UserId, a.Pswd).Scan(&name, &roleCode, &email); err != nil {
		errHandler.ErrMsg(w, err, 404)
		return
	}

	claims := jwt.MapClaims{
		"tokenType": "admin",
		"name":      name,
		"roleCode":  roleCode,
		"email":     email,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}

	tokenString, err := createToken(claims)
	if err != nil {
		errHandler.ErrMsg(w, err, 404)
		return
	}

	w.WriteHeader(http.StatusOK)
	var okReply OKReply
	okReply.Status = "OK"
	okReply.Token = tokenString
	json.NewEncoder(w).Encode(okReply)

}

func createToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}
