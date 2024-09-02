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

type LessonHeader struct {
	ModCode    string `json:"modCode"`
	LessonCode string `json:"lessonCode"`
	Title      string `json:"title"`
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

// -----------------------------------------------
// type LessonHeaders struct {
// 	LessonHeader []LessonHeader
// }

type LessonData struct {
	Pages []Page `json:"pages"`
}

type Page struct {
	Page  int8   `json:"page"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Step int8   `json:"step"`
	En   string `json:"En"`
	Id   string `json:"Id"`
}

//-----------------------------------------------

// Admin

func GetLesson(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonId string) {

	var strLessonData string

	if err := conn.QueryRow(context.Background(), "select lessonData from lesson where lessonCode = $1", lessonId).Scan(&strLessonData); err != nil {
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

func SyncApp(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {

	fileName := ""
	fileNames := []string{}

	appId, ok := r.Context().Value("appId").(string)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = "appId not found in SycnApp"
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	rows, err := conn.Query(context.Background(), "select fileName from sync where updatedOn >= (Select lastSync from app where appId  = $1) and substring(fileName,2,4) <= ( select lessonCode from progress where appId = $2 order by lessonCode desc limit 1)", appId, appId)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	defer rows.Close()

	for rows.Next() {

		if err := rows.Scan(&fileName); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			var nokReply NOKReply
			nokReply.Status = "NOK"
			nokReply.Errors = err.Error()
			json.NewEncoder(w).Encode(nokReply)
			return
		}

		fileNames = append(fileNames, fileName)

	}

	updAppStmt := "Update app set lastSync = Now() where appId = $1"

	_, err2 := conn.Exec(context.Background(), updAppStmt, appId)

	if checkError(w, err2) {
		return
	}

	type OKReply struct {
		Status    string
		FileNames []string
	}

	w.WriteHeader(http.StatusOK)
	var okReply OKReply
	okReply.Status = "OK"
	okReply.FileNames = fileNames
	json.NewEncoder(w).Encode(okReply)

}

func GetQuiz(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, quizCode string) {

	var strQuizData string

	if err := conn.QueryRow(context.Background(), "select quizData from quiz where quizCode = $1", quizCode).Scan(&strQuizData); err != nil {
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

func UpdQuiz(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, quizCode string) {

	quizData := QuizData{
		Quizes: []Quiz{},
	}

	json.NewDecoder(r.Body).Decode(&quizData)

	updAppStmt := "Update quiz set quizData = $1 where quizCode = $2"

	_, err := conn.Exec(context.Background(), updAppStmt, quizData, quizCode)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Quiz Updated"
	okReply.Message = fmt.Sprintf("Quiz %s Updated", quizCode)
	json.NewEncoder(w).Encode(okReply)

}

func UpdLesson(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonCode string) {

	lessonData := LessonData{
		Pages: []Page{},
	}

	json.NewDecoder(r.Body).Decode(&lessonData)

	updAppStmt := "Update lesson set lessonData = $1 where lessonCode = $2"

	_, err := conn.Exec(context.Background(), updAppStmt, lessonData, lessonCode)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Lesson Updated"
	okReply.Message = fmt.Sprintf("Lesson %s Updated", lessonCode)
	json.NewEncoder(w).Encode(okReply)

}

func UpdHeader(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, modCode string, lessonCode string) {

	lessonHeader := LessonHeader{}

	json.NewDecoder(r.Body).Decode(&lessonHeader)

	updAppStmt := "Update lesson set modCode = $1, lessonCode = $2, title = $3 where modCode = $4 and lessonCode = $5"

	_, err := conn.Exec(context.Background(), updAppStmt, lessonHeader.ModCode, lessonHeader.LessonCode, lessonHeader.Title, modCode, lessonCode)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Lesson Updated"
	okReply.Message = fmt.Sprintf("Modular %s : Lesson %s Updated", modCode, lessonCode)
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

// Application
func UpdProgress(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonCode string, result string) {

	appId, ok := r.Context().Value("appId").(string)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = "appId not found in claims"
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	//updAppStmtold := "Update app set lessonCode = $1 where appId = $2"

	updAppStmt := "insert into progress (appId, lessonCode, result) values ($1, $2, $3) ON CONFLICT(appId, lessonCode) DO UPDATE SET result = EXCLUDED.result"

	_, err := conn.Exec(context.Background(), updAppStmt, appId, lessonCode, result)

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = fmt.Sprintf("AppId %s Lesson Progress Updated", appId)
	json.NewEncoder(w).Encode(okReply)

}

// func UpdAppProgress(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {

// 	lesson := Lesson{}
// 	lesson.Lesson = "1"
// 	lesson.Page = "1"
// 	lesson.Page = "0%"

// 	progress := Progress{
// 		Done: []Lesson{},
// 	}
// 	appId, ok := r.Context().Value("appId").(string)

// 	if !ok {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		var nokReply NOKReply
// 		nokReply.Status = "NOK"
// 		nokReply.Errors = "appId not found in claims"
// 		json.NewEncoder(w).Encode(nokReply)
// 		return
// 	}

// 	json.NewDecoder(r.Body).Decode(&progress)

// 	updAppStmt := "Update app set progress = $1 where appId = $2"

// 	_, err := conn.Exec(context.Background(), updAppStmt, progress, appId)

// 	if checkError(w, err) {
// 		return
// 	}

// 	var okReply OKReply
// 	okReply.Status = "OK"
// 	okReply.Message = fmt.Sprintf("AppId %s Progress Updated", appId)
// 	json.NewEncoder(w).Encode(okReply)

// }

func RegisterApp(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {

	type NewApp struct {
		AppId       string `json:"appId"`
		Email       string `json:"email"`
		DeviceOs    string `json:"deviceOs"`
		NativeLingo string `json:"nativeLingo"`
	}

	var a NewApp

	lesson := Lesson{}
	lesson.Lesson = "1"
	lesson.Page = "1"
	lesson.Page = "0%"

	// progress := Progress{
	// 	Done: []Lesson{},
	// }

	//lastSync :=
	json.NewDecoder(r.Body).Decode(&a)

	insAppStmt := "insert into app (appId, email, active, deviceOs, nativeLingo, pmtLevel,  lastSync) values($1, $2, $3, $4, $5, $6, $7)"

	_, err := conn.Exec(context.Background(), insAppStmt, a.AppId, a.Email, "1", a.DeviceOs, a.NativeLingo, 0, "2024-08-24")

	if checkError(w, err) {
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Application Added"
	json.NewEncoder(w).Encode(okReply)

}

func GetProgress(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {

	var strLessonCode string
	var strResult string

	// type Lessons struct {
	// 	Lesson []Lesson
	// }

	type Lesson struct {
		LessonCode string
		Result     string
	}

	appId, ok := r.Context().Value("appId").(string)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = "AppId not found"
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	rows, err := conn.Query(context.Background(), "select lessonCode, result from progress where appId  = $1 order by lessonCode", appId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	defer rows.Close()

	lessons := []Lesson{}

	for rows.Next() {

		if err := rows.Scan(&strLessonCode, &strResult); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			var nokReply NOKReply
			nokReply.Status = "NOK"
			nokReply.Errors = err.Error()
			json.NewEncoder(w).Encode(nokReply)
			return
		}

		lesson := Lesson{}
		lesson.LessonCode = strLessonCode
		lesson.Result = strResult
		lessons = append(lessons, lesson)

	}

	type OKReply struct {
		Status  string
		Lessons []Lesson
	}

	w.WriteHeader(http.StatusOK)
	var okReply OKReply
	okReply.Status = "OK1"
	okReply.Lessons = lessons
	json.NewEncoder(w).Encode(okReply)

}

func GetHeaders(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, modCode string) {
	var strLessonCode string
	var strTitle string

	// 	rows, err := conn.Query("SELECT ename, sal FROM emp order by sal desc")
	//    if err != nil {
	//             panic(err)
	//    }

	// rows, err := conn.Query(context.Background(), "select lessonCode, title from lesson where modCode  = $1 order by lessonCode", modCode)
	rows, err := conn.Query(context.Background(), "select lessonCode, title from lesson where substring(lessonCode,1,2)  = $1 order by lessonCode", modCode)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var nokReply NOKReply
		nokReply.Status = "NOK"
		nokReply.Errors = err.Error()
		json.NewEncoder(w).Encode(nokReply)
		return
	}

	defer rows.Close()

	lessonHeaders := []LessonHeader{}

	for rows.Next() {

		if err := rows.Scan(&strLessonCode, &strTitle); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			var nokReply NOKReply
			nokReply.Status = "NOK"
			nokReply.Errors = err.Error()
			json.NewEncoder(w).Encode(nokReply)
			return
		}

		lessonHeader := LessonHeader{}
		lessonHeader.LessonCode = strLessonCode
		lessonHeader.Title = strTitle
		lessonHeaders = append(lessonHeaders, lessonHeader)

	}

	// bytLessonHeaders := []byte(strLessonHeaders)
	// var jsonLessonHeaders LessonHeaders

	// err := json.Unmarshal(bytLessonHeaders, &lessonHeaders)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	type OKReply struct {
		Status        string
		LessonHeaders []LessonHeader
	}

	w.WriteHeader(http.StatusOK)
	var okReply OKReply
	okReply.Status = "OK"
	okReply.LessonHeaders = lessonHeaders
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

// type claimskey int

// var claimsKey claimskey

// func JWTClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {

// 	claimsKey = 1
// 	claims, ok := ctx.Value(claimsKey).(jwt.MapClaims)
// 	return claims, ok
// }
