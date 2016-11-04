package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"net/http"
	"html/template"

	"git.ishopex.cn/matrix/gatling/lib"
)

var (
	AKey = "matrix"
	Secret = "i3jc810dkm4"
	OrderURL = "http://elastic-jst.ishopex.cn/cloud_order-*/_search"
	TaobaoRpcURL = "http://elastic-jst.ishopex.cn/rpc_log-*/_search"
	OterPlatformURL = "http://elastic-sh.ishopex.cn/rpc_log-*/_search"
	Method = "store.logistics.offline.send"
	Url = flag.String("mgo", "localhost:27017", "mgourl like : 127.0.0.1:27017")
	DB = flag.String("db", "elk", "mgo db name")
	Collection = flag.String("c", "record", "mgo collection")
	l = flag.String("log", "running.log", "logfile")
	LastTime string
	FORMATTIME = "2006-01-02T15:04:05.000Z"
	STANDARDTIEM = "2006-01-02 15:04:05"
)

type Series struct {
	Time  string    `json:"time"`
	Value []float64 `json:"value"`
}

type RpcSeries struct {
	Time  string `json:"time"`
	Value []RpcResult `json:"value"`
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

type RpcResult struct {
	Key   string `json:"key"`
	Count float64  `json:"doc_count"`
}

//获取淘宝log信息
func TaoBaoLog(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	if LastTime == "" {
		LastTime = time.Unix(now.Unix() - 5 * 60, 0).UTC().Format(FORMATTIME)
	}

	var (
		s Series
	)

	s.Time = time.Now().Format("15:04:05")
	Last_time, value := ComposeRes(OrderURL, LastTime, time.Now().UTC().Format(FORMATTIME))
	s.Value = value
	SaveResult(*Collection, s) //保存查询数据{时间:""，data:[总数,失败数]}
	LastTime = Last_time
	body, err := json.Marshal(s)
	if err != nil {
		log.Printf("[ERROR]页面数据出错 :%s", err)
		return
	}
	w.Write(body)
}

func Rpc(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
         var (
		 syncflag bool
		 uri  string
		 ty    string  //淘宝或者other
		 flag  string //同步或异步
	 )

	ty = r.PostForm.Get("type")
	flag = r.PostForm.Get("sync")

	if  ty == "" || flag == ""{
		log.Printf("[Crash]ajax post data is vaild ")
		return
	}

	switch flag {
	case "true":
		syncflag = true
	case "false":
	default:
		log.Printf("[Crash]ajax post data is vaild ")
		return
	}

	switch ty {
	case "taobao":
		uri = TaobaoRpcURL
	case "rpc":
		uri = OterPlatformURL
	default:
		log.Printf("[Crash]ajax post data is vaild ")
		return
	}


	now := time.Now().UTC()
	if LastTime == "" {
		LastTime = time.Unix(now.Unix() - 5 * 60, 0).UTC().Format(FORMATTIME)
	}
	t,rr:= RpcData(syncflag,uri,LastTime,now.Format(FORMATTIME))
	LastTime = t

	var s RpcSeries

	s.Time = time.Now().Format(STANDARDTIEM)
	s.Value = rr
	body ,err := json.Marshal(s)
	if err != nil {
		log.Printf("[ERROR]页面数据出错 :%s", err)
		return
	}

	SaveResult(ty+flag,s) //保存数据
	w.Write(body)
}

//跳转首页
func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Printf("[ERROR]加载主页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

func Platform(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("platform.html")
	if err != nil {
		log.Printf("[ERROR]加载其他平台页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

//
func  Total(w http.ResponseWriter, r *http.Request)  {
	t, err := template.ParseFiles("test.html")
	if err != nil {
		log.Printf("[ERROR]加载rpc页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}
func  taobaos(w http.ResponseWriter, r *http.Request)  {
	t, err := template.ParseFiles("taobaosync.html")
	if err != nil {
		log.Printf("[ERROR]加载taobaosync页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}
func  taobaoa(w http.ResponseWriter, r *http.Request)  {
	t, err := template.ParseFiles("taobaoasync.html")
	if err != nil {
		log.Printf("[ERROR]加载taobaoaysnc页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}
func  rpcs(w http.ResponseWriter, r *http.Request)  {
	t, err := template.ParseFiles("rpcsync.html")
	if err != nil {
		log.Printf("[ERROR]加载rpcs页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

func  rpca(w http.ResponseWriter, r *http.Request)  {
	t, err := template.ParseFiles("rpcasync.html")
	if err != nil {
		log.Printf("[ERROR]加载rpc页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}



//从mongodb中获取之前存储的数据
func historydata(w http.ResponseWriter, r *http.Request) {
	e, err := QueryData()
	if err != nil {
		log.Printf("[ERROR]加载历史数据出错:%s", err)
	}
	body, err := json.Marshal(e)
	if err != nil {
		log.Printf("[ERROR]加载历史数据出错 :%s", err)
		return
	}
	w.Write(body)
}

//rpc历史数据
func rpchistorydata(w http.ResponseWriter,r *http.Request)  {
	r.ParseForm()
	e, err := QueryRpcData(r.PostForm.Get("type")+r.PostForm.Get("sync"))
	if err != nil {
		log.Printf("[ERROR]加载rpc历史数据出错:%s", err)
	}
	body, err := json.Marshal(e)
	if err != nil {
		log.Printf("[ERROR]加载rpc历史数据出错 :%s", err)
		return
	}
	w.Write(body)
}

func main() {
	flag.Parse()
	fw, err := os.OpenFile(*l, os.O_CREATE | os.O_RDWR, 0664)
	defer fw.Close()
	if err != nil {
		log.Printf("[ERROR]logfile open error:%s", err)
		os.Exit(-1)
	}
	log.SetOutput(fw)
	log.SetFlags(log.Lshortfile | log.Ltime)
	err = lib.StartMongo("mongodb://" + *Url + "/")
	if err != nil {
		os.Exit(-1)
	}
	log.Println("[INFO]START.........................")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/elk_monitor/taobao", TaoBaoLog)
	http.HandleFunc("/elk_monitor/historydata", historydata)
	http.HandleFunc("/index", index)
	http.HandleFunc("/taobaosync",taobaos)
	http.HandleFunc("/taobaoasync",taobaoa)
	http.HandleFunc("/rpcsync",rpcs)
	http.HandleFunc("/rpcasync",rpca)
	http.HandleFunc("/platform", Platform)
	http.HandleFunc("/rpc/realtime", Rpc)
	http.HandleFunc("/total",Total)
	http.HandleFunc("/rpc/historydata",rpchistorydata)
	http.ListenAndServe(":8889", nil)
}
