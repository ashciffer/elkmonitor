package main

import (
	"log"
	"time"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

//syncflag 为真 取同步数据
func ComposeRpcLog(syncflag bool,start,end string) string {
	ti := time.Now().Unix()
	mm := Sign_str(Secret, strconv.Itoa(int(ti)))
	s := Sec{
		Key:    AKey,
		Value:  mm,
		Number: ti,
	}

	p := Params{}
	p.Auth = s
	must1 := map[string]interface{}{
		"match": map[string]string{
			"type": "async_request",
		},
	}
	must2 := map[string]interface{}{
		"match": map[string]interface{}{
			"method": "store.logistics.offline.send",
		},
	}
	must3 := map[string]interface{}{
		"range": map[string]interface{}{
			"@timestamp": map[string]interface{}{
				"gte": start,
				"lt":  end,
			},
		},
	}

	if syncflag {
		must1["match"].(map[string]string)["type"] = "sync_request"
	}
	must := []map[string]interface{}{must1,must2, must3}
	query := map[string]interface{}{
		"sort": map[string]interface{}{
			"time": map[string]string{
				"order": "desc",
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":must,
			},
		},
		"from": 0,
		"size": 1,
		"aggs":map[string]interface{}{
			"status":map[string]interface{}{
				"terms":map[string]interface{}{
					"field":"status.raw",
					"size":0,
					"shard_size":0,
				},
			},
		},
	}

	p.Query = query

	body, err := json.Marshal(p)
	if err != nil {
		log.Printf("[ERROR]query is error:%s", err)
		return ""
	}
	return string(body)
}

//
func  RpcData(flag bool,uri,start ,end string)(lasttime string,rr []RpcResult){
	var (
		d map[string]interface{}
		err error
		resp *http.Response
		b []byte
		s RpcResult
	)
	lasttime = end
	//捕捉panic
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("[Crash]Rpc response error :%s",rec)
		}
	}()
	query := ComposeRpcLog(flag,start,end)
         resp,err = Request(uri,"POST",query)
	if err != nil {
		log.Printf("[ERROR]Request is erorr:%s", err)
		return
	}
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR]Resp body read  erorr:%s", err)
		return
	}
	//log.Printf("[DEBUG]resp : %s\n", string(b))

	err = json.Unmarshal(b, &d)
	if err != nil {
		return
	}

	hits := d["hits"].(map[string]interface{})
	if len(hits["hits"].([]interface{})) > 0 {
		lasttime = hits["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})["@timestamp"].(string)
	}

	buckets := d["aggregations"].(map[string]interface{})["status"].(map[string]interface{})["buckets"].([]interface{})
	if len(buckets) <=0 {
            rr = append(rr,RpcResult{
		    Key:"total",
		    Count:0.0,
	    })
	}else{
		for _,v := range  buckets{
			s.Key = v.(map[string]interface{})["key"].(string)
			if s.Key == ""{
				s.Key = "total"
			}
			s.Count = v.(map[string]interface{})["doc_count"].(float64)
			rr =append(rr,s)
		}
	}

	return
}