package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4"
)

var (
	dsn string
	cnt int64
)

type OKReply struct {
	Status  string
	Message string
}

type NOKReply struct {
	Status string
	Errors string
}

type AppId struct {
	AppId string `json:"appid"`
	Pin   string `json:"pin"`
}

type Lesson struct {
	Lesson string `json:"Lesson"`
	Page   string `json:"Page"`
	Result string `json:"Result"`
}

type Progress struct {
	Done []Lesson `json:"Done"`
}

// type Done struct {
// 	Done []Lesson `json:"Done"`
// }

// func UpdData(conn *sql.DB) {

// 	// update
// 	updateStmt := `update "students" set "name"=$1, "roll_number"=$2 where "roll_number"=$3`
// 	_, e := conn.Exec(updateStmt, "Rachel", 24, 24)
// 	checkError(e)

// }

func InfoApp(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, appId string) {

	var strNativeLingo, strDeviceOs, strProgress string
	var bolActive bool

	if err := conn.QueryRow(context.Background(), "select active, nativeLingo, deviceOs, progress from apps where appId = $1", appId).Scan(&bolActive, &strNativeLingo, &strDeviceOs, &strProgress); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	bytProgress := []byte(strProgress)
	var jsonProgress Progress

	err := json.Unmarshal(bytProgress, &jsonProgress)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	type OKReply struct {
		Status      string
		Active      bool
		NativeLingo string
		DeviceOs    string
		Progress    Progress
	}

	w.WriteHeader(http.StatusOK)
	var okReply OKReply
	okReply.Status = "OK"
	okReply.DeviceOs = strDeviceOs
	okReply.NativeLingo = strNativeLingo
	okReply.Active = bolActive
	okReply.Progress = jsonProgress
	json.NewEncoder(w).Encode(okReply)

}

func UpdProgress(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, appId string) {

	lesson := Lesson{}
	lesson.Lesson = "1"
	lesson.Page = "1"
	lesson.Page = "0%"

	progress := Progress{
		Done: []Lesson{},
	}

	json.NewDecoder(r.Body).Decode(&progress)

	updAppStmt := "Update apps set progress = $1 where appId = $2"

	_, err := conn.Exec(context.Background(), updAppStmt, progress, appId)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Application Updated"
	json.NewEncoder(w).Encode(okReply)

}

// func InsertApps(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, appId string , pin string, deviceOs string, nativ ) {
// dynamic
func AddApp(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, appId string) {

	type NewAppId struct {
		Pin         string `json:"pin"`
		DeviceOs    string `json:"deviceOs"`
		NativeLingo string `json:"nativeLingo"`
	}

	var a NewAppId

	lesson := Lesson{}
	lesson.Lesson = "1"
	lesson.Page = "1"
	lesson.Page = "0%"

	progress := Progress{
		Done: []Lesson{},
	}

	json.NewDecoder(r.Body).Decode(&a)

	insAppStmt := "insert into apps (appId, pin, active, deviceOs, nativeLingo, progress) values($1, $2, $3, $4, $5, $6)"

	_, err := conn.Exec(context.Background(), insAppStmt, appId, a.Pin, "1", a.DeviceOs, a.NativeLingo, progress)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Application Added"
	json.NewEncoder(w).Encode(okReply)

}

// func ListData(conn *sql.DB) {
// 	rows, err := conn.Query(`SELECT "name", "roll_number" FROM "students"`)
// 	checkError(err)

// 	defer rows.Close()
// 	for rows.Next() {
// 		var name string
// 		var roll_number int

// 		err = rows.Scan(&name, &roll_number)
// 		checkError(err)

// 		fmt.Println(name, roll_number)
// 	}

// }

func checkError(w http.ResponseWriter, err error) bool {
	if err != nil {
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
		return true
	} else {
		return false
	}
}

func openDB() (*sql.DB, error) {

	connStr := "postgres://postgres:mysecretpassword@localhost/db_1?sslmode=disable"

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	//defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil

}

func ConnectToDB() *sql.DB {
	//dsn := os.Getenv("DSN")

	for {
		connection, err := openDB()

		if err != nil {
			log.Println("Postgress not yet ready...")
			cnt++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if cnt > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for 2 seconds...")
		time.Sleep(2 * time.Second)
		continue
	}

}
