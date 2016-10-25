package main

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type Params struct {
	Auth  Sec                    `json:"auth"`
	Query map[string]interface{} `json:"query"`
}

type Sec struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Number int64  `json:"number"`
}

func TestMem(t *testing.T) {
	ti := time.Now().Unix()
	mm := Sign_str(Secret, strconv.Itoa(int(ti)))
	s := Sec{
		Key:    "matrix",
		Value:  mm,
		Number: ti,
	}

	p := Params{}
	p.Auth = s
	// must1 := map[string]interface{}{
	// 	"match": map[string]interface{}{
	// 		"result_status": "fail",
	// 	},
	// }
	must2 := map[string]interface{}{
		"match": map[string]interface{}{
			"step": "result",
		},
	}
	must3 := map[string]interface{}{
		"range": map[string]interface{}{
			"@timestamp": map[string]string{
				"gte": "2016-10-25T04:04:57.780Z",
				"lt":  "2016-10-25T04:09:57.780Z",
			},
		},
	}

	query := map[string]interface{}{
		"sort": map[string]interface{}{
			"time": map[string]string{
				"order": "desc",
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					//must1,
					must2,
					must3,
				},
			},
		},
		"from": 0,
		"size": 1,
	}

	p.Query = query

	body, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := Request(OrderURL, "POST", string(body))
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(b))
	var result map[string]interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(reflect.TypeOf(result["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})["@timestamp"]))
}
