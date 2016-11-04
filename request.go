package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.ishopex.cn/matrix/gatling/lib"
)


//请求
func Request(url string, method, params string) (*http.Response, error) {
	client := &http.Client{
		Transport: http.DefaultTransport,
	}

	body := strings.NewReader(params)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	return resp, err
}

//拆分数据
func GetTotalAndTime(data []byte) (time string, total float64, err error) {
	var (
		d map[string]interface{}
	)

	defer func(err *error, t float64, ti string) {
		if rec := recover(); rec != nil {
			*err = errors.New(fmt.Sprintf("%s", rec))

		}
	}(&err, total, time)

	err = json.Unmarshal(data, &d)
	if err != nil {
		return "", 0, err
	}

	hits := d["hits"].(map[string]interface{})
	total = hits["total"].(float64)
	time = hits["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})["@timestamp"].(string)

	return time, total, err
}

//生成query条件 f为真则query为求总数条件
func ComposeQuery(f bool, start, end string) string {
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
		"match": map[string]interface{}{
			"result_status": "fail",
		},
	}
	must2 := map[string]interface{}{
		"match": map[string]interface{}{
			"step": "result",
		},
	}
	must3 := map[string]interface{}{
		"range": map[string]interface{}{
			"@timestamp": map[string]string{
				"gte": start,
				"lt":  end,
			},
		},
	}

	must := []map[string]interface{}{must2, must3}
	if !f {
		must = append(must, must1)
	}
	query := map[string]interface{}{
		"sort": map[string]interface{}{
			"time": map[string]string{
				"order": "desc",
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
		"from": 0,
		"size": 1,
	}

	p.Query = query

	body, err := json.Marshal(p)
	if err != nil {
		log.Printf("[ERROR]query is error:%s", err)
		return ""
	}
	return string(body)
}

//获取总数
func GetTotal(uri string, f bool, start, end string) (string, float64) {
	//log.Printf("[DEBUG]start - %s,end - %s", start, end)
	query := ComposeQuery(f, start, end)
	resp, err := Request(uri, "POST", query)
	if err != nil {
		log.Printf("[ERROR]Request is erorr:%s", err)
		return start, 0
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR]Resp body read  erorr:%s", err)
		return start, 0
	}
	//log.Printf("[DEBUG]resp : %s", string(b))
	t, value, err := GetTotalAndTime(b)
	if err != nil {
		log.Printf("[ERROR]result is error:%s", err)
		return start, 0
	}
	return t, value
}


//保存数据
func SaveResult(col string, e interface{}) {
	m := lib.Mongo.Clone()
	err := m.DB(*DB).C(col).Insert(e)
	if err != nil {
		log.Printf("[ERROR]数据保存出错:%s", err)
	}
}

//取数据query()
func QueryData() ([]Series, error) {
	e := []Series{}
	m := lib.Mongo.Clone()
	err := m.DB(*DB).C(*Collection).Find(nil).Sort("-_id").Limit(12).All(&e)
	var k int
	var succ []Series
	var Re   []Series
	for _, v := range e {
		if v.Value != nil {
			succ = append(succ, v)
		}
	}

	for i := 0;i<(12-len(succ));i++{
		tmp := Series{Time:"00:00:"+strconv.Itoa(k),Value:[]float64{0,0}}
		Re = append(Re,tmp)
		k++
	}

	for n := len(succ); n > 0; n-- {
		Re = append(Re, succ[n - 1])
	}
	return Re, err
}

func QueryRpcData(col string) ([]RpcSeries, error) {
	e := []RpcSeries{}
	var re []RpcSeries

	m := lib.Mongo.Clone()
	err := m.DB(*DB).C(col).Find(nil).Limit(12).Sort("-_id").All(&e)
	var succ []RpcSeries
	for _, v := range e {
		if v.Value != nil {
			succ = append(succ, v)
		}
	}
	re =append(re,succ...)

	tmp_result := RpcResult{
		Key:"total",
		Count:0,
	}

	for i := (12-len(succ));i>0;i--{
		tmp := RpcSeries{Time:"2006 00:00:"+strconv.Itoa(i),Value:[]RpcResult{tmp_result}}
		re = append(re,tmp)
	}
	return re, err
}

//返回值{FORMATTIME,[total,fail]}
func ComposeRes(uri, start, end string) (last_time string, res []float64) {
	_, total := GetTotal(uri, false, start, end)
	fail_time, fail := GetTotal(uri, true, start, end)
	last_time = fail_time
	res = append(res, total, fail)
	return
}


