package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"html/template"
	"net/http"

	"github.com/robfig/cron"
	"gopkg.in/mgo.v2"
)

var (
	AKey            = "****"
	Secret          = "*****"
	OrderURL        = "*"
	shanghaiURL     = "*"
	TaobaoRpcURL    = "*"
	OterPlatformURL = "*"
	Method          = "store.logistics.offline.send"
	Mgo             *mgo.Session
	Url             = flag.String("mgo", "192.168.75.54:27017,192.168.75.54:27027,192.168.75.54:27037", "mgourl like : 127.0.0.1:27017")
	DB              = flag.String("db", "elk", "mgo db name")
	l               = flag.String("log", "running.log", "logfile")
	spec            = flag.String("cron", "@every 5m", "cron spec")
	FORMATTIME      = "2006-01-02T15:04:05.000Z"
	STANDARDTIEM    = "2006-01-02 15:04:05"
)

var (
	ts_lt string
	ta_lt string
	rs_lt string
	ra_lt string
)

type Series struct {
	Time  string    `json:"time"`
	Value []float64 `json:"value"`
}

type RpcSeries struct {
	Time  string      `json:"time"`
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
	Key   string  `json:"key"`
	Count float64 `json:"doc_count"`
}

//获取淘宝log信息
func TaoBaoLog(ty string) {
	var (
		uri string
	)

	if ty == "taobao" {
		uri = OrderURL
	} else {
		uri = shanghaiURL
	}

	last_do_time, err := GetCronTime(ty)
	if err != nil {
		last_do_time = time.Unix(time.Now().UTC().Unix()-15*60, 0).UTC().Format(FORMATTIME)
		SetCronTime(ty)
	}

	var (
		s Series
	)

	s.Time = time.Now().Format(STANDARDTIEM)
	t, value := ComposeRes(uri, last_do_time, time.Now().UTC().Format(FORMATTIME))
	s.Value = value

	err = ModiyLastDoTime(ty, t)
	if err != nil {
		log.Println("[Error]修改 %s 的最后修改时间失败：%s", ty, err)
	}
	SaveResult(ty, s) //保存查询数据{时间:""，data:[总数,失败数]}
}

//上海发货 数据分拣
func Rpc(c string) {

	switch c {
	case "taobaotrue":
		rpchandler(true, TaobaoRpcURL, c)
	case "taobaofalse":
		rpchandler(false, TaobaoRpcURL, c)
	case "rpctrue":
		rpchandler(true, OterPlatformURL, c)
	case "rpcfalse":
		rpchandler(false, OterPlatformURL, c)
	default:
		log.Printf("[Crash]ajax post data is vaild ")
		return
	}
}

func rpchandler(syncflag bool, uri, col string) {
	now := time.Now().UTC()
	last_do_time, err := GetCronTime(col)
	if err != nil {
		last_do_time = time.Unix(time.Now().UTC().Unix()-15*60, 0).UTC().Format(FORMATTIME)
		SetCronTime(col)
	}

	t, rr := RpcData(syncflag, uri, last_do_time, now.Format(FORMATTIME))

	var s RpcSeries

	s.Time = time.Now().Format(STANDARDTIEM)
	s.Value = rr
	err = ModiyLastDoTime(col, t)
	if err != nil {
		log.Println("[Error]修改 %s 的最后修改时间失败：%s", col, err)
	}

	SaveResult(col, s) //保存数据
}

//跳转首页
func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/pages/index.html")
	if err != nil {
		log.Printf("[ERROR]加载主页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

func Platform(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/pages/platform.html")
	if err != nil {
		log.Printf("[ERROR]加载其他平台页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

//
func Total(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/pages/test.html")
	if err != nil {
		log.Printf("[ERROR]加载rpc页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}
func taobaos(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/pages/taobaosync.html")
	if err != nil {
		log.Printf("[ERROR]加载taobaosync页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}
func taobaoa(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/pages/taobaoasync.html")
	if err != nil {
		log.Printf("[ERROR]加载taobaoaysnc页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}
func rpcs(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/pages/rpcsync.html")
	if err != nil {
		log.Printf("[ERROR]加载rpcs页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

func rpca(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/pages/rpcasync.html")
	if err != nil {
		log.Printf("[ERROR]加载rpc页面出错:%s", err)
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

//定时任务
func Job() {
	go TaoBaoLog("taobao")
	go TaoBaoLog("shanghai")
	go Rpc("taobaotrue")
	go Rpc("taobaofalse")
	go Rpc("rpctrue")
	go Rpc("rpcfalse")
}

//从mongodb中获取之前存储的数据
func historydata(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var ty string
	ty = r.PostForm.Get("type")
	if ty == "" {
		log.Printf("[Crash]ajax post data is invaild ")
		return
	}

	e, err := QueryData(ty)
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
func rpchistorydata(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	e, err := QueryRpcData(r.PostForm.Get("type") + r.PostForm.Get("sync"))
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
	cronapp := cron.New()
	cronapp.AddFunc(*spec, Job)
	cronapp.Start()

	fw, err := os.OpenFile(*l, os.O_CREATE|os.O_RDWR, 0664)
	defer fw.Close()
	if err != nil {
		log.Printf("[ERROR]logfile open error:%s", err)
		os.Exit(-1)
	}
	//log.SetOutput(fw)
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	Mgo, err = mgo.Dial("mongodb://" + *Url + "/")
	if err != nil {
		os.Exit(-1)
	}

	log.Println("[INFO]START.........................")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/elk_monitor/historydata", historydata)
	http.HandleFunc("/index", index)
	http.HandleFunc("/taobaosync", taobaos)
	http.HandleFunc("/taobaoasync", taobaoa)
	http.HandleFunc("/rpcsync", rpcs)
	http.HandleFunc("/rpcasync", rpca)
	http.HandleFunc("/platform", Platform)
	//http.HandleFunc("/rpc/realtime", Rpc)
	http.HandleFunc("/total", Total)
	http.HandleFunc("/rpc/historydata", rpchistorydata)
	err = http.ListenAndServe(":8889", nil)
	if err != nil {
		log.Println("httpserver can't establish ： %s ", err)
	}
}
