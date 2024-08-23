package auth

import (
	"context"
	"encoding/json"

	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
)

const adminTokenSecret = "Hello, Sysbit"

var (
	secretKey = []byte("secret-key")
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
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
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
	roleCode := ""

	if err := conn.QueryRow(context.Background(), "select name, roleCode, email from login where loginCode = $1 and pswd = $2", a.UserId, a.Pswd).Scan(&name, &roleCode, &email); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
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

// func ChkAppToken(h http.Handler, w http.ResponseWriter, r *http.Request) error {
// 	tokenString := r.Header.Get("Authorization")

// 	if tokenString == "" {
// 		//w.WriteHeader(http.StatusUnauthorized)
// 		return errors.New("Missing Authorization Header")
// 	}

// 	tokenString = tokenString[len("Bearer "):]

// 	// vars := mux.Vars(r)
// 	// appId := vars["appId"]

// 	err := verifyToken(tokenString, r)
// 	if err != nil {
// 		//w.WriteHeader(http.StatusUnauthorized)
// 		//return errors.New("Invalid Token")
// 		return err

// 	}

// 	return nil
// }

// func ChkAdminToken(w http.ResponseWriter, r *http.Request) error {
// 	tokenString := r.Header.Get("Authorization")

// 	if tokenString == "" {
// 		//w.WriteHeader(http.StatusUnauthorized)
// 		return errors.New("Missing Authorization Header")
// 	}

// 	tokenString = tokenString[len("Bearer "):]

// 	err := verifyToken(tokenString, r)
// 	if err != nil {
// 		//w.WriteHeader(http.StatusUnauthorized)
// 		//return errors.New("Invalid Token")
// 		return err

// 	}

// 	return nil
// }

func createToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// jwt.MapClaims{
	// 	"appId": appId,
	// 	"exp":   time.Now().Add(time.Hour * 24).Unix(),
	// }
	//claims

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

// func verifyToken(tokenString string, r *http.Request) error {
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		return secretKey, nil
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	if !token.Valid {
// 		return fmt.Errorf("invalid token")
// 	}

// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok {
// 		return fmt.Errorf("Unable to extract claims")
// 	}

// 	// if claims["appId"].(string) != appId {
// 	// 	return fmt.Errorf("Token does not match appID")
// 	// }

// 	//fmt.Println(claims["appId"].(string))

// 	//r = r.WithContext(SetJWTClaimsContext(r.Context(), claims))

// 	//newCtx := context.WithValue(r.Context(), "userId", "userId")

// 	r.Header.Add("AppId", claims["appId"].(string))

// 	//r = r.WithContext(newCtx)

// 	return nil
// }

// type claimskey int

// var claimsKey claimskey

// func SetJWTClaimsContext(ctx context.Context, claims jwt.MapClaims) context.Context {

// 	claimsKey = 1
// 	return context.WithValue(ctx, claimsKey, claims)
// }
