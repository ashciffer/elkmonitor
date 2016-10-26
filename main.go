package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"git.ishopex.cn/matrix/gatling/lib"
)

import "net/http"
import "html/template"

var (
	AKey       = "matrix"
	Secret     = "i3jc810dkm4"
	OrderURL   = "http://elastic-jst.ishopex.cn/cloud_order-*/_search"
	Method     = "store.logistics.offline.send"
	Url        = flag.String("mgo", "localhost:27017", "mgourl like : 127.0.0.1:27017")
	DB         = flag.String("db", "elk", "mgo db name")
	Collection = flag.String("c", "record", "mgo collection")
	LastTime   string
	FORMATTIME = "2006-01-02T15:04:05.000Z"
)

type Series struct {
	Time  string    `json:"time"`
	Value []float64 `json:"value"`
}

type Params struct {
	Auth  Sec                    `json:"auth"`
	Query map[string]interface{} `json:"query"`
}

type Sec struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Number int64  `json:"number"`
}

func elk_monitor(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	if LastTime == "" {
		LastTime = time.Unix(now.Unix()-5*60, 0).UTC().Format(FORMATTIME)
	}

	var (
		s     Series
		start string
		total float64
		fail  float64
	)

	s.Time = time.Now().Format("15:04:05")
	_, total = GetTotal(true, LastTime, now.Format(FORMATTIME))
	start, fail = GetTotal(false, LastTime, now.Format(FORMATTIME))
	s.Value = []float64{total, fail}
	LastTime = start
	body, err := json.Marshal(s)
	if err != nil {
		log.Printf("[ERROR]页面数据出错 :%s", err)
		return
	}
	w.Write(body)
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Println("[ERROR]加载主页面出错:%s", err)
	}
	t.Execute(w, nil)
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ltime)
	err := lib.StartMongo("mongodb://" + *Url + "/")
	if err != nil {
		os.Exit(-1)
	}
	log.Println("[INFO]START.........................")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/elk_monitor/data", elk_monitor)
	http.HandleFunc("/index", index)
	http.ListenAndServe(":8889", nil)
}
