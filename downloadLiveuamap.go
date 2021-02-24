package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func DownloadLiveuamap(c chan int){
	const dataFolder string ="liveuamapdata"
	res,err :=	http.Get("http://167.172.51.95:3000/liveuamap/json")
	if err !=nil{
		fmt.Println("Could not resolve")
		c<- -1 //let the waiting channel know we failed ): lest we crash the application
		return
	}
	defer res.Body.Close()
	resp,err := ioutil.ReadAll(res.Body)
	if err !=nil{
		fmt.Println(err.Error())
	}
	var data struct {
		All []string
	}
	err = json.Unmarshal(resp,&data)
	if err !=nil{
		fmt.Println(err.Error())
	}
	for _,jurl := range data.All{
		name := strings.Split(jurl,"//")[1]
		actualName:= strings.Split(name,"/")[2]
		err:= os.Mkdir("./"+dataFolder+"/"+actualName, os.ModePerm)
		if err !=nil{
			fmt.Println(err.Error())
			//if already exists then no need to try that one
			continue //skips this iteration as no need
		}
		out,err := os.Create("./"+dataFolder+"/"+actualName+"/0all.json")
		if err !=nil{
			fmt.Println("Cannot create file")
		}
		resp,err := http.Get(jurl)
		n,err := io.Copy(out,resp.Body)
		if err !=nil{
			fmt.Println(err.Error())
		}
		out.Close()
		fmt.Println(n)
		fmt.Println(actualName)
	}
	c<-1 //let the caller know its done, we were successful in downloading the live data
}
