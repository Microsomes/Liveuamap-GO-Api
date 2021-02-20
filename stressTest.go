package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func sendPing(url string,ch chan<-string){

 		data := map[string]string{"Title": "from stress test", "Desc": "Desc hello"}
		jsonValue, _ := json.Marshal(data)

		respBody, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
 		if err != nil {
			fmt.Println("Error")
			fmt.Println(err)
			ch<- "error"
			return
		}

		respBody.Body.Close()



	ch<- "done"
}

func main(){

	localURL:= "http://localhost:10000/create"
	networkURL:= "http://139.162.193.140:10000/create"

	fmt.Println(localURL,networkURL)

	c:= make(chan string)

		for {
 			go sendPing(localURL, c)
			time.Sleep(1 * time.Millisecond)
			fmt.Println("--------")
 			fmt.Println("Took")
 			fmt.Println(<-c)
		}


}
