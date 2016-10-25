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

//Mongo存储结构
type Elk struct {
	time int64
	data string
}

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
func GetTotalAndTime(data []byte) (string, float64, error) {
	var (
		d   map[string]interface{}
		err error
	)

	defer func(err *error) {
		if rec := recover(); rec != nil {
			*err = errors.New(fmt.Sprintf("%s", rec))
		}
	}(&err)

	err = json.Unmarshal(data, &d)
	if err != nil {
		return "", 0, err
	}

	hits := d["hits"].(map[string]interface{})
	total := hits["total"].(float64)
	time := hits["hits"].([]interface{})[0].(map[string]interface{})["@timestamp"].(string)
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
		log.Printf("[error]query is error:%s", err)
		return ""
	}
	return string(body)
}

//获取总数
func GetTotal(f bool, start, end string) (string, float64) {
	query := ComposeQuery(f, start, end)
	resp, err := Request(OrderURL, "POST", query)
	if err != nil {
		return start, 0
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return start, 0
	}
	SaveResult(b)
	t, value, err := GetTotalAndTime(b)
	if err != nil {
		return start, 0
	}
	return t, value
}

//保存数据
func SaveResult(body []byte) {
	m := lib.Mongo.Clone()
	var e Elk
	e.time = time.Now().Unix()
	e.data = string(body)
	err := m.DB(*DB).C(*Collection).Insert(e)
	if err != nil {
		log.Printf("[ERROR]数据保存出错:%s", err)
	}
}
