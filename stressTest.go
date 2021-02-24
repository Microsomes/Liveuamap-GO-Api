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


func stess(){

	localURL:= "http://localhost:10000/create"
	networkURL:= "http://139.162.193.140:10000/create"

	fmt.Println(localURL,networkURL)

	c:= make(chan string)

	totalMinutes:=10

	fmt.Println(totalMinutes)

	totalReq:=1000

	totalTimeTaken:=time.Now()

	for {
		if totalReq<=0{
			break
		}
		totalReq--
		fmt.Println("Left:",totalReq)
		startMinute:= time.Now()

		for i:=0;i<100;i++{
			go sendPing(networkURL, c)
			fmt.Println("--------")
			fmt.Println("Took")
			fmt.Println(<-c)
			duration:= time.Now().Sub(startMinute)
			totalTimeTaken.Add(duration)
			fmt.Println("Took",duration)

		}



	}
}

func main(){

	c:= make(chan int)
 	for i:=0;i<100;i++{
 		go stess()
	}

	<-c
}
