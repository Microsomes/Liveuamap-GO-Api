package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type MaepLog struct {
	Id      int       `json:Id`
	Title   string    `json:"Title"`
	Desc    string    `json:Desc`
	Created time.Time `json:Created`
}

type ErrorStruct struct {
	Status string `json:"Status"`
	Msg    string `json:"Msg"`
}

var Maeps []MaepLog

func homePage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	fmt.Fprint(w, "GO")
}

func returnAllMaepsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type AllCollection struct {
		Total int       `json:Total`
		Data  []MaepLog `json:Data`
	}
	toReturn := AllCollection{
		Total: len(Maeps),
		Data:  Maeps,
	}
	json.NewEncoder(w).Encode(toReturn)
}

func handlerRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/create", createNewMaep).Methods("POST")
	myRouter.HandleFunc("/all", returnAllMaepsHandler)
	myRouter.HandleFunc("/item/{id}", returnSingleMaep)
	myRouter.HandleFunc("/update/{id}", updateMaep).Methods("PUT")
	myRouter.HandleFunc("/total",returnTotal)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func returnTotal(w http.ResponseWriter, r*http.Request){
	defer r.Body.Close()

	total := len(Maeps)
	fmt.Fprint(w,total)
}

func returnSingleMaep(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	var key = vars["id"]
	id, err := strconv.Atoi(key)
	if err != nil {
		repl := ErrorStruct{Status: "ERR", Msg: "Not a number, [0-9]+ "}
		json.NewEncoder(w).Encode(repl)
		return
	}
	for _, maep := range Maeps {
		if maep.Id == id {
			json.NewEncoder(w).Encode(maep)
			return
		}
	}
	repl := ErrorStruct{Status: "ERR", Msg: "Not found"}
	json.NewEncoder(w).Encode(repl)
}

func updateMaep(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	var key = vars["id"]
	id, err := strconv.Atoi(key)
	if err != nil {
		json.NewEncoder(w).Encode(&ErrorStruct{
			Status: "ERR",
			Msg:    "Not a number, [0-9]+ ",
		})
		return
	}

	//the id is bigger then what exists
	if id > len(Maeps) {
		json.NewEncoder(w).Encode(&ErrorStruct{
			Status: "ERR",
			Msg:    "Maep ID is out of bounds, aka it doesnt exist",
		})
		return
	}

	inQuestion := Maeps[id]

	resBody, err := ioutil.ReadAll(r.Body)

	if err != nil {

		json.NewEncoder(w).Encode(&ErrorStruct{
			Status: "ERR",
			Msg:    "Something wrong with the json body",
		})
		return
	}

	var newData MaepLog
	json.Unmarshal(resBody, &newData)

	fmt.Println("new:", newData)

	fmt.Println("old:", inQuestion)

	//replace old with new

	Maeps[id] = newData

	json.NewEncoder(w).Encode(&ErrorStruct{
		Status: "OK",
		Msg:    "Updated",
	})

}

func createNewMaep(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		json.NewEncoder(w).Encode(&ErrorStruct{
			Status: "ERR",
			Msg:    "Something wrong with the json body",
		})
	}

	var maep MaepLog
	json.Unmarshal(reqBody, &maep)
	var lastID int = Maeps[len(Maeps)-1].Id
	maep.Id = lastID + 1
	maep.Created = time.Now()
	Maeps = append(Maeps, maep)
	fmt.Println(maep)

	fmt.Fprint(w, lastID+1)


}

func main() {
	Maeps = []MaepLog{
		MaepLog{Title: "Trump elected again", Desc: "Republicans move to nominate trump again", Id: 0},
	}
	handlerRequests()
}
