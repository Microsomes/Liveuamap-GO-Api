package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
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
	myRouter.HandleFunc("/loaderio-7a46e6ccb1dbbc0fe1eca0f5848e3d8c",handlerLoaderio)


	//liveuamap section
	myRouter.HandleFunc("/liveuamap",getSchemaLiveuamap)
	myRouter.HandleFunc("/liveuamap/total",getTotalRecordsLiveUa)
	myRouter.HandleFunc("/liveuamap/getdates",getAllLiveDates)
	myRouter.HandleFunc("/liveuamap/getdate/{date}",getDataFromDateLiveUa)
	myRouter.HandleFunc("/liveuamap/getdate/{date}/{pageNo}",getDataFromDateLiveUaPagi)
	myRouter.HandleFunc("/liveuamap/gettags",getAllTags)
	myRouter.HandleFunc("/liveuamap/getdate/{date}/{tag}",getDataByTagUsingDate)
	myRouter.HandleFunc("/liveuamap/post/{id}",getLiveUamapItem)

	//search functionalities
	myRouter.HandleFunc("/liveuamap/searchtitle/{query}",searchLiveuamapSearchText)

	myRouter.HandleFunc("/liveuamap/install",installlive)

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}



func extractQuery(r *http.Request) string {
	vars := mux.Vars(r)
	var querystr= vars["query"]
	return  querystr
}

func searchLiveuamapSearchText(w http.ResponseWriter, r*http.Request){
	query:=extractQuery(r)
	db:= dbConn()
	defer db.Close()
	selDB,err:= db.Query("SELECT id,Title,PostDate FROM liveuamap_v1 WHERE Title like ?","%"+query+"%")
	if err !=nil{
		fmt.Fprint(w,"Error searching")
	}
	type postR struct {
		PostID int
		Title string
		PostDate string
		MoreDetails string
	}
	var allPosts []postR
	for selDB.Next(){
		var id int
		var title string
		var PostDate string
		selDB.Scan(&id,&title,&PostDate)
		allPosts=append(allPosts,postR{
			PostID: id,
			Title: title,
			PostDate: PostDate,
			MoreDetails: "/liveuamap/post/"+strconv.Itoa(id),
		})
	}
	type stdReply struct {
		Status string
		Msg string
		Total int
		Data []postR
	}
	json.NewEncoder(w).Encode(&stdReply{
		Status: "OK",
		Msg: "All Titles with their id, more details can be available at the more details end point",
		Total: len(allPosts),
		Data: allPosts,
	})
 }


func getMetaLiveuamap(metaID int,c chan interface{}){
	db:=dbConn()
	defer db.Close()
	res:= db.QueryRow("SELECT * FROM liveuamap_meta_v1 WHERE id=?",metaID)
	var id int
	var keym string
	var valm string
	err:= res.Scan(&id,&keym,&valm)
	if err!=nil{
		c<- ErrorStruct{}
	}
	//it worked
	type stdMeta struct {
		Id int
		Keym string
		Valm string
	}
	c<-stdMeta{
		Id: id,
		Keym: keym,
		Valm: valm,
	}
 }

func getLiveUamapItem(w http.ResponseWriter, r*http.Request){
	id:= extractID(r)
	db:= dbConn()

	var selDB *sql.Row=  db.QueryRow("SELECT * FROM liveuamap_v1 WHERE id=?",id)
	var postid int
	var title string
	var postImage string
	var source string
	var postDate string
	var timeAgo string
	var postLink string
	var postIcon string
	var metaID int
	err:= selDB.Scan(&postid, &title,&postImage,&source,&postDate,&timeAgo,&postLink,&postIcon,&metaID)

	c:= make(chan interface{})

	if reflect.TypeOf(c).Name() =="ErrorStruct"{
		json.NewEncoder(w).Encode(&ErrorStruct{
			Status: "ERR",
			Msg: "Invalid meta id",
		})
		return
	}
	go getMetaLiveuamap(metaID,c)
	//grabs the meta id
	metaData:=<-c


	fmt.Println(metaData)


	if err!=nil{
		json.NewEncoder(w).Encode(&ErrorStruct{
			Status: "ERR",
			Msg: err.Error(),
		})
	}

	type stdReply struct {
		Id int
		Title string
		PostImage string
		Source string
		PostDate string
		TimeAgo string
		PostLink string
		PostIcon string
		MetaID int
		MetaData interface{}
	}
	json.NewEncoder(w).Encode(&stdReply{
		Id: postid,
		Title: title,
		PostImage: postImage,
		Source: source,
		PostDate: postDate,
		TimeAgo: timeAgo,
		PostLink: postLink,
		PostIcon: postIcon,
		MetaID: metaID,
		MetaData: metaData,
	})
	defer db.Close()
}

func getSchemaLiveuamap(w http.ResponseWriter, r*http.Request){
	var data =map[string]string{
		"Status":"ok",
		"/searchtitle/trump":"<- searches for the database title for the word trump returns the title and the post id",
		"/getdates":"<-returns all the current dates in system",
		"/install":"<-dumps all records and re-imports new news data, ps dont spam this. Will be removed later",
		"/getdate/01-2-2021":"<-Returns the post of a specified date, example date added, you can get all the available dates from /getdates",
		"/post/2":"<- returns post that is of the id 2, now you can try guess the id you want or just use getdate to grab ids of posts of a certain id",
	}
	json.NewEncoder(w).Encode(data)
}


func getTotalRecordsLiveUa(w http.ResponseWriter, r*http.Request){
	db:= dbConn()
   var dbRow  *sql.Row =	db.QueryRow("SELECT count(id) as 'total' FROM liveuamap_v1")
   var totalRecords int
   err:=dbRow.Scan(&totalRecords)
   if err!=nil{
   	//error found
   	fmt.Println(err.Error())
   	json.NewEncoder(w).Encode(&ErrorStruct{
   		Status: "ERR",
   		Msg: "Qurey failed to run, cannot grab totals count. ",
	})
   	return
   }

   type stdReply struct{
   	Status string
   	Msg string
   	Total int
   }
   json.NewEncoder(w).Encode(&stdReply{
   	Status: "OK",
   	Msg: "<-Total records of liveuamap posts captured by the api service",
   	Total: totalRecords,
   })
	defer db.Close()
}


func setUpImport(db *sql.DB, a chan int){
	//we drop both tables
	//todo check for errors for time sake ill skip this for now
	//todo- when u have time check
	 db.Exec("DROP TABLE liveuamap_v1")
	 db.Exec("DROP TABLE liveuamap_meta_v1")
	_,err := db.Exec("CREATE TABLE liveuamap_v1  (id integer NOT NULL AUTO_INCREMENT PRIMARY KEY, " +
		"Title text," +
		"PostImage text," +
		"Source varchar(255)," +
		"PostDate varchar(10),"+
		"TimeAgo varchar(20)," +
		"PostLink varchar(255)," +
		"PostIcon text," +
		"MetaID integer," +
		"INDEX (PostDate)) ENGINE=InnoDB DEFAULT CHARSET=utf8 DEFAULT COLLATE utf8_unicode_ci")

	_,err = db.Exec("CREATE TABLE liveuamap_meta_v1  (id integer NOT NULL AUTO_INCREMENT PRIMARY KEY," +
		"Keym varchar(255)," +
		"Valm varchar(255)," +
		"INDEX (Keym)) ENGINE=InnoDB DEFAULT CHARSET=utf8 DEFAULT COLLATE utf8_unicode_ci")

	if err !=nil{
		fmt.Println(err)
		//table probably already exists which is fine
	}

	//once all the database tables are created lets load the io file data to be imported

	files,err := ioutil.ReadDir("./liveuamapdata")

	if err !=nil{
		fmt.Println("Ohh no not working ")
	}

	for _,f := range files{
		dirName :=f.Name()
		jsonB,err := ioutil.ReadFile("./liveuamapdata/"+dirName+"/0all.json")
		if err !=nil{
			fmt.Println("Opps something wrong reading the data")
		}

		type meta struct{
			Coord string
			Tags []string
		}
		type postItem struct{
			Title string  `json:"title"`
			PostImage string
			Source string
			PostDate string
			TimeAgo string
			PostLink string
			PostIcon string
			Meta meta
		}
		var data struct{
			TotalPosts  int `json:"TotalPosts"`
			Posts []postItem  `json:"posts"`
		}

		json.Unmarshal(jsonB, &data)


		for _,d:= range data.Posts{

			insQueer, err := db.Prepare("INSERT INTO liveuamap_v1(Title," +
				"PostImage," +
				"Source," +
				"PostDate," +
				"TimeAgo," +
				"PostLink," +
				"PostIcon," +
				"MetaID)  VALUES(?,?,?,?,?,?,?,?)")

			insMetaQueer,err := db.Prepare("INSERT INTO liveuamap_meta_v1 (Keym,Valm) VALUES (?,?)")

			if err!=nil{
				fmt.Println(err.Error())
			}

			res,err:=insMetaQueer.Exec("coordinate",d.Meta.Coord)
			if err !=nil{
				fmt.Println(err.Error())
			}
			insMetaQueer.Close()
			metaID,err := res.LastInsertId()
			if err !=nil {
				fmt.Println(err.Error())
			}

			allMonths := [12]string{
				"January",
				"February",
				"March",
				"April",
				"May",
				"June",
				"July",
				"August",
				"September",
				"October",
				"November",
				"December",
			}

			timeSplit:= strings.Split(d.PostDate," ")
			tm:= timeSplit[1]
			dm:= timeSplit[0]
			ym:= timeSplit[2]

			mint:= -1

			for i,d:= range allMonths{
				if strings.ToUpper(d)== strings.ToUpper(tm) {
					fmt.Println("Found Date")
					mint=i
				}
			}

			var toMonth= strconv.Itoa(mint+1)


			toDate:= dm+"/"+ toMonth + "/"+ ym

			//s := fmt.Sprintf("%06d", 12) // returns '000012' as a String


			fmt.Println(mint)

			insQueer.Exec(d.Title,
				d.PostImage,
				d.Source,
				toDate,
				d.TimeAgo,
				d.PostLink,
				d.PostIcon,
				metaID)


			if err !=nil{
				fmt.Println(err)
				fmt.Println("Their was an error inserting")

			}else {

				err = insQueer.Close()
				if err != nil {
					fmt.Println(err)
				}
				//fmt.Println(insQueer)
			}
		}



	}

	a<-9


}

func dbConn()*sql.DB{
	db_host:=os.Getenv("DB_HOST")
 	db_username := os.Getenv("DB_USERNAME")
 	db_password := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")
	connectionString:= db_username+":"+db_password+"@tcp(127.0.0.1:3306)/"+db_name
 	fmt.Println(db_host,db_username,db_password)
	var db, err = sql.Open("mysql", connectionString)
	if err !=nil{
		panic(err.Error())
	}
	return db
}


var isRunning bool =false

func installlive(w http.ResponseWriter, r*http.Request){
 	fmt.Println("Go MySQL Tutorial")
	// Open up our database connection.
	// I've set up a database on my local machine using phpmyadmin.
	// The database is called testDb

	var c = make(chan int)
	go DownloadLiveuamap(c)

	fmt.Println(<-c)

	//if n !=-2{
	//	fmt.Println("Book proceed")
	//}else{
	//	//if the scrapper does not resolve theirs no point trying to continue the import as
	//	//were just gonna reimport whats already in our cache/db
	//	fmt.Fprint(w,"The scrapper is down, contact tayyan54@Gmail.com")
	//	return
	//}
	//prevents a double import since its just gonna mess up
	//the database inserts and cause a chance for duplicate records
	//on install we usually drop the table and recreate it,
	//image doing double the work lol
	//itll be fihgting so i use this to check and prevent
	//a call during a job
	if isRunning{
			fmt.Fprint(w,"Already installing")
			return
		}
		isRunning=true
		db:= dbConn()
		c = make(chan int)
		go setUpImport(db, c)
		fmt.Println(<-c)
		isRunning=false
		fmt.Fprint(w,"Installed to live db")
}




func handlerLoaderio(w http.ResponseWriter, r*http.Request){
	defer r.Body.Close()
	fmt.Fprint(w,"loaderio-7a46e6ccb1dbbc0fe1eca0f5848e3d8c\n")
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



func removeDuplicates(arr []string) []string{
		words_string := map[string]bool{}

		for i := range arr {
			words_string[arr[i]]=true
		}
	desiredOutput := []string{}  // Keep all keys from the map into a slice.

		for j,_ := range words_string{
			desiredOutput= append(desiredOutput,j)
		}
		return desiredOutput

}



func getDataByTagUsingDate(w http.ResponseWriter, r * http.Request){

}

func getAllTags(w http.ResponseWriter, r* http.Request){
	defer r.Body.Close()

	//hard coded 1 date, will optimise later
	fileData,err := ioutil.ReadFile("./liveuamapdata/Friday_February_12th_2021/0all.json")

	if err !=nil{
		fmt.Println("Error")
		json.NewEncoder(w).Encode(&ErrorStruct{
			Status: "Err",
			Msg: "Ops something went wrong, contact tayyan54@gmail.com",
		})
		return
	}

	type meta struct{
 		Tags []string
	}
	type postItem struct{
		Meta meta

	}
	var data struct{
		TotalPosts  int `json:"TotalPosts"`
		Posts []postItem  `json:"posts"`
	}

	var err2 = json.Unmarshal(fileData,&data)

	if err2!=nil{
		json.NewEncoder(w).Encode(&ErrorStruct{
			Status: "Err",
			Msg: "Something went wrong parsing the data, contact tayyan54@gmail.com",
		})
		return
	}

	// I KNOW THIS NEEDS TO BE CLEANER LOL ITS NOT OPITMIZED!!! I COME FROM JAVASRIPT SORRRRRRRRRRY

	var allTags []string

	for _,l := range data.Posts{
		fmt.Println(l.Meta.Tags)
		for _,tag := range l.Meta.Tags{
			allTags= append(allTags,tag)
		}
	}


	var cleanList []string

	cleanList= removeDuplicates(allTags)




	json.NewEncoder(w).Encode(cleanList)





}

func getDateLive(date string, page int, db *sql.DB, c chan interface{}){
	var maxlimit int= 10
	//always only returns 10 posts hard coded for now
	var offsetBy int= page*maxlimit
	selDB,err := db.Query("SELECT id FROM liveuamap_v1 WHERE PostDate=? order by id asc LIMIT ? OFFSET ?",date,10,offsetBy)
	if err !=nil{
 		c<- ErrorStruct{
 			Status: "ERR",
 			Msg: "Erm something wrong with the database getDateLive",
		}
		return
	}
	defer selDB.Next()
	defer db.Close()
	var AllPostIds []int
	//all post ids will be stored here
	for selDB.Next(){
		var id int
		err := selDB.Scan(&id)
		if err!=nil{
 			c<- ErrorStruct{
 				Status: "ERR",
 				Msg: "Something went wrong querying the row getDateLive",
			}
			break
		}
		AllPostIds= append(AllPostIds,id)
	}
	c<- AllPostIds
 }

 func extractDateFromURL(r*http.Request) string {
	 vars :=mux.Vars(r)
	 var id= vars["date"]
	 var dateSplit= strings.Split(id,"-")
	 var newDate string
	 for i,d := range dateSplit{
		 if i==len(dateSplit)-1{
			 newDate += d
		 }else {
			 newDate += d + "/"
		 }
	 }
	 return newDate
 }


func extractID(r *http.Request) int {
	vars := mux.Vars(r)
	var pageNo= vars["id"]
	p,err :=strconv.Atoi(pageNo)
	if err!=nil{
		//todo better logging needs to be added to catch this error
		fmt.Println("Error extracting pageNo")
		return 0
	}
	return  p
}

 func extractPageNo(r *http.Request) int {
 	vars := mux.Vars(r)
 	var pageNo= vars["pageNo"]
 	p,err :=strconv.Atoi(pageNo)
 	if err!=nil{
 		//todo better logging needs to be added to catch this error
 		fmt.Println("Error extracting pageNo")
 		return 0
	}
 	return  p
 }

 func getDataFromDateLiveUaPagi(w http.ResponseWriter, r*http.Request){
 	newDate:= extractDateFromURL(r)
	pageNo := extractPageNo(r)
	fmt.Println(pageNo)
 	 db:= dbConn()
	 c:= make(chan interface{})
	 go getDateLive(newDate,pageNo,db,c)
	 var receiv = <-c
	 if reflect.TypeOf(receiv).Name()=="ErrorStruct"{
		 fmt.Fprint(w,"Something went wrong with the request, please try again or contact the support")
	 }else {
		 type stdReply struct {
			 Status string
			 Msg string
			 PageNo int
			 Ids interface{}
		 }
		 json.NewEncoder(w).Encode(&stdReply{
			 Status: "OK",
			 Msg: "<- List of the recent post ids. Pagination /1 <- you can request page 10 by apending /10",
			 PageNo: pageNo,
			 Ids:receiv,
		 })
	 }
 }

func getDataFromDateLiveUa(w http.ResponseWriter, r*http.Request){
	defer r.Body.Close()
	var newDate string= extractDateFromURL(r)
	db:= dbConn()
	c:= make(chan interface{})
	go getDateLive(newDate,0,db,c)
	var receiv = <-c
	if reflect.TypeOf(receiv).Name()=="ErrorStruct"{
		fmt.Fprint(w,"Something went wrong with the request, please try again or contact the support")
	}else {
		//todo breaking dry poliicy need to extract this to be reusable as its used twice
		type stdReply struct {
			Status string
			Msg string
			PageNo int
			Ids interface{}
		}
		json.NewEncoder(w).Encode(&stdReply{
			Status: "OK",
			Msg: "<- List of the recent post ids. Pagination /1 <- you can request page 10 by apending /10",
			PageNo: 0,
			Ids:receiv,
		})
	}
}

var allGetDatesCache  []MaepletCacheDate

//todo improve the cache//cache for get date
type MaepletCacheDate struct{
	key string
	valm stdReply_getAllLiveDates
	Expires time.Time
}
func (c MaepletCacheDate) set(duration time.Time){
	c.Expires=duration//sets the duration of the cache
	allGetDatesCache= append(allGetDatesCache,c)
}


type toItem_getAllLiveDates struct{
	Date string
	Total int
}

type stdReply_getAllLiveDates struct {
	Status string
	IsCache bool
	CacheExp string
	Msg string
	Data []toItem_getAllLiveDates
}

func getAllLiveDates(w http.ResponseWriter, r*http.Request){
	defer r.Body.Close()

	for _,d:= range allGetDatesCache{
		if d.key== r.RequestURI{
			fmt.Println("found cache")

			//check if the cache has expired

			if time.Until(d.Expires).Seconds() >= time.Second.Seconds(){
				d.valm.IsCache=true
				d.valm.CacheExp= d.Expires.String()
				json.NewEncoder(w).Encode(d.valm)
				return
			}else{
				//cache has expired
				fmt.Println("cache expired")
			}
		}
	}

	db:= dbConn()
	selDb,err := db.Query("SELECT DISTINCT PostDate as 'pd', (select count(id) from liveuamap_v1 WHERE PostDate=pd) as 'total'  from liveuamap_v1")
	if err !=nil{
		json.NewEncoder(w).Encode(&ErrorStruct{
			Status: "ERR",
			Msg: "Error grabbing details",
		})
		return
	}

	var totalItems []toItem_getAllLiveDates
	for selDb.Next() {
		var date string
		var total int
 		err = selDb.Scan(&date,&total)
 		if err != nil{
 			panic(err.Error())
		}
		totalItems= append(totalItems,toItem_getAllLiveDates{
			Date: date,
			Total: total,
		})
 	}
	defer db.Close()
	repl:= &stdReply_getAllLiveDates{
		Status: "OK",
		IsCache: false,
		CacheExp:"n/a",
		Msg: "Returns the dates of the post and the total in db, its not ordered, thats on the todo list give us a few weeks to get that sorted. If its not fixed bug me _22/02/2021",
		Data: totalItems,
	}
	//lets cache this
	cacheSet:= MaepletCacheDate{
		key: r.RequestURI,
		valm: *repl,
	}
	//cache of exactly 12 hours
	expiry:= time.Now().Add(time.Hour*12)
	//when the cache expires
	cacheSet.set(expiry)
	json.NewEncoder(w).Encode(repl)
}


func main() {
	//set up db connection
	os.Setenv("DB_HOST","localhost")
	os.Setenv("DB_PORT","3306")
	os.Setenv("DB_USERNAME","root")
	os.Setenv("DB_PASSWORD","root")
	os.Setenv("DB_NAME","maeplet")

	Maeps = []MaepLog{
		MaepLog{Title: "Trump elected again", Desc: "Republicans move to nominate trump again", Id: 0},
	}
	handlerRequests()
}
