package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func sendPing(){

 		data := map[string]string{"Title": "from stress test", "Desc": "Desc hello"}
		jsonValue, _ := json.Marshal(data)

		_, err := http.Post("http://139.162.193.140:10000/create", "application/json", bytes.NewBuffer(jsonValue))

		if err != nil {
			fmt.Println("Error")
		}

}

func main(){


	c:= make(chan int)


	for {
		fmt.Println("Sending Request")
		go sendPing()
		time.Sleep(1 * time.Millisecond)
	}

	<-c



}
