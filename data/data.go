package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sysbitBroker/errHandler"

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

type LessonDetails struct {
	LessonDetails []LessonDetail `json:"lessonDetail"`
}

type LessonDetail struct {
	StepCode string `json:"stepCode"`
	SpeakTxt string `json:"speakTxt"`
}

type LessonHeader struct {
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
	Nbr   int    `json:"nbr"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Step int    `json:"step"`
	En   string `json:"En"`
	Fx   string `json:"Fx"`
}

//-----------------------------------------------

// Admin

// func GetLesson(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonId string) {

// 	var strLessonData string

// 	if err := conn.QueryRow(context.Background(), "select lessonData from lesson where lessonCode = $1", lessonId).Scan(&strLessonData); err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		var nokReply NOKReply
// 		nokReply.Status = "NOK"
// 		nokReply.Errors = err.Error()
// 		json.NewEncoder(w).Encode(nokReply)
// 		return
// 	}

// 	bytLessonData := []byte(strLessonData)
// 	var jsonLessonData LessonData

// 	err := json.Unmarshal(bytLessonData, &jsonLessonData)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	type OKReply struct {
// 		Status     string
// 		LessonData LessonData
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	var okReply OKReply
// 	okReply.Status = "OK"
// 	okReply.LessonData = jsonLessonData
// 	json.NewEncoder(w).Encode(okReply)

// }

func GetLessonDetail(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonCode string, lingoCode string) {

	var pageCode, stepCode, en, fx string

	//rows, err := conn.Query(context.Background(), "select pageCode, stepCode, lingoCode, conversation from lessonDetail where lessonCode = $1 order by pageCode, stepCode, lingoCode ", lessonCode)
	rows, err := conn.Query(context.Background(), "select A.pagecode, A.stepcode, A.speakTxt as en , b.speakTxt as fx from lessonDetail A left join lessonDetail B on a.lessonCode = b.lessonCode and a.pageCode = b.pageCode and a.stepCode = b.stepCode where A.lessonCode = $1 and A.lingoCode = 'en' and b.lingoCode = $2 order by a.pageCode, a.stepCode", lessonCode, lingoCode)

	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var page Page
	var step Step

	lessonData := LessonData{}
	lessonData.Pages = []Page{}
	lastPageCode := ""

	stepNbr := 0

	for rows.Next() {

		if err := rows.Scan(&pageCode, &stepCode, &en, &fx); err != nil {
			errHandler.ErrMsg(w, err, http.StatusInternalServerError)
			return
		}

		if pageCode != lastPageCode {
			page = Page{}
			page.Nbr, _ = strconv.Atoi(pageCode)
			page.Steps = []Step{}
			lastPageCode = pageCode
			lessonData.Pages = append(lessonData.Pages, page)

		}

		step = Step{}
		stepNbr += 1
		step.Step, _ = strconv.Atoi(stepCode)
		step.En = en
		step.Fx = fx

		lessonData.Pages[len(lessonData.Pages)-1].Steps = append(lessonData.Pages[len(lessonData.Pages)-1].Steps, step)

	}

	type OKReply struct {
		Status     string
		LessonData LessonData
	}

	w.WriteHeader(http.StatusOK)
	var okReply OKReply
	okReply.Status = "OK"
	okReply.LessonData = lessonData
	json.NewEncoder(w).Encode(okReply)

}

func SyncApp(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {

	fileName := ""
	fileNames := []string{}

	appId, ok := r.Context().Value("appId").(string)

	if !ok {
		errHandler.ErrMsg(w, errors.New("application id not in claims"), http.StatusInternalServerError)
		return
	}

	rows, err := conn.Query(context.Background(), "select fileName from sync where updatedOn >= (Select lastSync from app where appId  = $1) and substring(fileName,2,4) <= ( select lessonCode from progress where appId = $2 order by lessonCode desc limit 1)", appId, appId)

	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	for rows.Next() {

		if err := rows.Scan(&fileName); err != nil {
			errHandler.ErrMsg(w, err, http.StatusInternalServerError)
			return
		}

		fileNames = append(fileNames, fileName)

	}

	updAppStmt := "Update app set lastSync = Now() where appId = $1"

	_, err2 := conn.Exec(context.Background(), updAppStmt, appId)

	if err2 != nil {
		errHandler.ErrMsg(w, err2, http.StatusInternalServerError)
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

func GetQuizDetail(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, quizCode string, lingoCode string) {

	var strQuizData string

	if err := conn.QueryRow(context.Background(), "select quizData from quiz where quizCode = $1 and lingoCode = $2", quizCode, lingoCode).Scan(&strQuizData); err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	bytQuizData := []byte(strQuizData)
	var jsonQuizData QuizData

	err := json.Unmarshal(bytQuizData, &jsonQuizData)
	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
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

func GetQuiz(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, quizCode string) {

	var strQuizData string

	if err := conn.QueryRow(context.Background(), "select quizData from quiz where quizCode = $1", quizCode).Scan(&strQuizData); err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	bytQuizData := []byte(strQuizData)
	var jsonQuizData QuizData

	err := json.Unmarshal(bytQuizData, &jsonQuizData)
	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
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

func UpdQuiz(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, quizCode string, lingoCode string) {

	quizData := QuizData{
		Quizes: []Quiz{},
	}

	json.NewDecoder(r.Body).Decode(&quizData)

	updAppStmt := "Update quiz set quizData = $1 where quizCode = $2 and lingoCode = $3"

	_, err := conn.Exec(context.Background(), updAppStmt, quizData, quizCode, lingoCode)

	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Quiz Updated"
	okReply.Message = fmt.Sprintf("Quiz %s Updated", quizCode)
	json.NewEncoder(w).Encode(okReply)

}

// func UpdLesson(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonCode string) {

// 	lessonData := LessonData{
// 		Pages: []Page{},
// 	}

// 	json.NewDecoder(r.Body).Decode(&lessonData)

// 	updAppStmt := "Update lesson set lessonData = $1 where lessonCode = $2"

// 	_, err := conn.Exec(context.Background(), updAppStmt, lessonData, lessonCode)

// 	if err != nil {
// 		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
// 		return
// 	}

// 	var okReply OKReply
// 	okReply.Status = "OK"
// 	okReply.Message = "Lesson Updated"
// 	okReply.Message = fmt.Sprintf("Lesson %s Updated", lessonCode)
// 	json.NewEncoder(w).Encode(okReply)

// }

func UpdLessonStep(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonCode string, lingoCode string) {

	lessonDetail := LessonDetail{}

	json.NewDecoder(r.Body).Decode(&lessonDetail)

	updAppStmt := "Update lessonDetail set speakTxt = $1 where lessonCode = $2  and stepCode = $3 and lingoCode = $4"

	_, err := conn.Exec(context.Background(), updAppStmt, lessonDetail.SpeakTxt, lessonCode, lessonDetail.StepCode, lingoCode)

	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	//okReply.Message = "Lesson Updated"
	okReply.Message = fmt.Sprintf("Lesson %s -  Step %s - Lingo %s  Updated", lessonCode, lessonDetail.StepCode, lingoCode)
	json.NewEncoder(w).Encode(okReply)

}

func UpdLessonHeader(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, modCode string, lessonCode string) {

	lessonHeader := LessonHeader{}

	json.NewDecoder(r.Body).Decode(&lessonHeader)

	updAppStmt := "Update lessonHeader set  lessonCode = $1, title = $2 where lessonCode = $3"

	_, err := conn.Exec(context.Background(), updAppStmt, lessonHeader.LessonCode, lessonHeader.Title, modCode, lessonCode)

	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = "Lesson Updated"
	okReply.Message = fmt.Sprintf("Modular %s : Lesson %s Updated", modCode, lessonCode)
	json.NewEncoder(w).Encode(okReply)

}

// func GetConv(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, appId string) {

// 	var strNativeLingo, strDeviceOs, strProgress string
// 	var bolActive bool

// 	if err := conn.QueryRow(context.Background(), "select active, nativeLingo, deviceOs, progress from app where appId = $1", appId).Scan(&bolActive, &strNativeLingo, &strDeviceOs, &strProgress); err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		var nokReply NOKReply
// 		nokReply.Status = "NOK"
// 		nokReply.Errors = err.Error()
// 		json.NewEncoder(w).Encode(nokReply)
// 		return
// 	}

// 	bytProgress := []byte(strProgress)
// 	var jsonProgress Progress

// 	err := json.Unmarshal(bytProgress, &jsonProgress)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	type OKReply struct {
// 		Status      string
// 		Active      bool
// 		NativeLingo string
// 		DeviceOs    string
// 		Progress    Progress
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	var okReply OKReply
// 	okReply.Status = "OK"
// 	okReply.DeviceOs = strDeviceOs
// 	okReply.NativeLingo = strNativeLingo
// 	okReply.Active = bolActive
// 	okReply.Progress = jsonProgress
// 	json.NewEncoder(w).Encode(okReply)

// }

// Application
func UpdProgress(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, lessonCode string, result string) {

	appId, ok := r.Context().Value("appId").(string)

	if !ok {
		errHandler.ErrMsg(w, errors.New("application id not in claims"), http.StatusInternalServerError)
		return
	}

	updAppStmt := "insert into progress (appId, lessonCode, result) values ($1, $2, $3) ON CONFLICT(appId, lessonCode) DO UPDATE SET result = EXCLUDED.result"

	_, err := conn.Exec(context.Background(), updAppStmt, appId, lessonCode, result)

	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	var okReply OKReply
	okReply.Status = "OK"
	okReply.Message = fmt.Sprintf("AppId %s Lesson Progress Updated", appId)
	json.NewEncoder(w).Encode(okReply)

}

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

	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
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
		errHandler.ErrMsg(w, errors.New("application id not in claims"), http.StatusInternalServerError)
		return
	}

	rows, err := conn.Query(context.Background(), "select lessonCode, result from progress where appId  = $1 order by lessonCode", appId)
	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	lessons := []Lesson{}

	for rows.Next() {

		if err := rows.Scan(&strLessonCode, &strResult); err != nil {
			errHandler.ErrMsg(w, err, http.StatusInternalServerError)
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

func GetLessonHeaders(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, modCode string) {
	var strLessonCode string
	var strTitle string

	rows, err := conn.Query(context.Background(), "select lessonCode, title from lessonHeader where substring(lessonCode,1,2)  = $1 order by lessonCode", modCode)

	if err != nil {
		errHandler.ErrMsg(w, err, http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	lessonHeaders := []LessonHeader{}

	for rows.Next() {

		if err := rows.Scan(&strLessonCode, &strTitle); err != nil {
			errHandler.ErrMsg(w, err, http.StatusInternalServerError)
			return
		}

		lessonHeader := LessonHeader{}
		lessonHeader.LessonCode = strLessonCode
		lessonHeader.Title = strTitle
		lessonHeaders = append(lessonHeaders, lessonHeader)

	}

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
