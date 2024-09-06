package errHandler

import (
	"encoding/json"
	"net/http"
)

type NOKReply struct {
	Status string
	Errors string
}

func ErrMsg(w http.ResponseWriter, err error, status int) {
	var nokReply NOKReply
	nokReply.Status = "NOK"
	nokReply.Errors = err.Error()
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(nokReply)

}
