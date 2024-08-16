package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v4"
)

// var (
// 	dsn string
// 	cnt int64
// )

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

// ----------------------------
type QuizData struct {
	Quizes []Quiz `json:"quizes"`
}

type Selection struct {
	Choice     string `json:"choice"`
	Desription string `json:"description"`
}

type Quiz struct {
	Nbr        int         `json:"nbr"`
	Context    string      `json:"context"`
	Question   string      `json:"question"`
	Selections []Selection `json:"selections"`
	Answer     string      `json:"answer"`
	Reason     string      `json:"reason"`
}

//-----------------------------------------------

type LessonData struct {
	Pages []Page `json:"pages"`
}

type Page struct {
	Nbr   int8   `json:"nbr"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Step int8   `json:"step"`
	En   string `json:"En"`
	Id   string `json:"Id"`
}

//-----------------------------------------------

func GetAppInfo(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, appId string) {

	var strNativeLingo, strDeviceOs, strProgress string
	var bolActive bool

	if err := conn.QueryRow(context.Background(), "select active, nativeLingo, deviceOs, progress from app where appId = $1", appId).Scan(&bolActive, &strNativeLingo, &strDeviceOs, &strProgress); err != nil {
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

func GetLesson(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonId string) {

	var strLessonData string

	if err := conn.QueryRow(context.Background(), "select lessonData from lesson where lessonId = $1", lessonId).Scan(&strLessonData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	bytLessonData := []byte(strLessonData)
	var jsonLessonData LessonData

	err := json.Unmarshal(bytLessonData, &jsonLessonData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	type OKReply struct {
		Status     string
		LessonData LessonData
	}

	w.WriteHeader(http.StatusOK)
	var okReply OKReply
	okReply.Status = "OK"
	okReply.LessonData = jsonLessonData
	json.NewEncoder(w).Encode(okReply)

}

func GetQuiz(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, quizId string) {

	var strQuizData string

	if err := conn.QueryRow(context.Background(), "select quizData from quiz where quizId = $1", quizId).Scan(&strQuizData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	bytQuizData := []byte(strQuizData)
	var jsonQuizData QuizData

	err := json.Unmarshal(bytQuizData, &jsonQuizData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	type OKReply struct {
		Status   string
		QuizData QuizData
	}

	w.WriteHeader(http.StatusOK)
	var okReply OKReply
	okReply.Status = "OK"
	okReply.QuizData = jsonQuizData
	json.NewEncoder(w).Encode(okReply)

}

func UpdQuiz(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, quizId string) {

	quizData := QuizData{
		Quizes: []Quiz{},
	}

	json.NewDecoder(r.Body).Decode(&quizData)

	updAppStmt := "Update quiz set quizData = $1 where quizId = $2"

	_, err := conn.Exec(context.Background(), updAppStmt, quizData, quizId)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Quiz Updated"
	json.NewEncoder(w).Encode(okReply)

}

func UpdLesson(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonId string) {

	lessonData := LessonData{
		Pages: []Page{},
	}

	json.NewDecoder(r.Body).Decode(&lessonData)

	updAppStmt := "Update lesson set lessonData = $1 where lessonId = $2"

	_, err := conn.Exec(context.Background(), updAppStmt, lessonData, lessonId)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Lesson Updated"
	json.NewEncoder(w).Encode(okReply)

}

func GetConv(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, appId string) {

	var strNativeLingo, strDeviceOs, strProgress string
	var bolActive bool

	if err := conn.QueryRow(context.Background(), "select active, nativeLingo, deviceOs, progress from app where appId = $1", appId).Scan(&bolActive, &strNativeLingo, &strDeviceOs, &strProgress); err != nil {
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

func UpdAppInfo(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, appId string) {

	lesson := Lesson{}
	lesson.Lesson = "1"
	lesson.Page = "1"
	lesson.Page = "0%"

	progress := Progress{
		Done: []Lesson{},
	}

	json.NewDecoder(r.Body).Decode(&progress)

	updAppStmt := "Update app set progress = $1 where appId = $2"

	_, err := conn.Exec(context.Background(), updAppStmt, progress, appId)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Application Updated"
	json.NewEncoder(w).Encode(okReply)

}

func RegisterApp(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, appId string) {

	type NewAppId struct {
		Email       string `json:"email"`
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

	insAppStmt := "insert into app (appId, email, active, deviceOs, nativeLingo, progress) values($1, $2, $3, $4, $5, $6)"

	_, err := conn.Exec(context.Background(), insAppStmt, appId, a.Email, "1", a.DeviceOs, a.NativeLingo, progress)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Application Added"
	json.NewEncoder(w).Encode(okReply)

}

// Generic Error Function

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
