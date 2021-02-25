package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main(){
	r,err:=http.Get("http://localhost:10000/liveuamap/getdates")
	if err!=nil{
		log.Fatal("Error")
	}
	respBody,err := ioutil.ReadAll(r.Body)
	if err!=nil{
		log.Fatal(err)
	}
	fmt.Println(string(respBody))
	time.Sleep(time.Hour)
}
