package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	//"reflect"

	"fmt"
	"strconv"
	"testing"
	"time"
)

func Mem(st, size int, must map[string]interface{}) (float64, map[string]interface{}) {
	ti := time.Now().Unix()
	mm := Sign_str(Secret, strconv.Itoa(int(ti)))
	s := Sec{
		Key:    "matrix",
		Value:  mm,
		Number: ti,
	}

	p := Params{}
	p.Auth = s

	query := map[string]interface{}{
		"sort": map[string]interface{}{
			"time": map[string]string{
				"order": "asc",
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					must,
				},
			},
		},
		"from": st,
		"size": size,
		"aggs": map[string]interface{}{
			"status": map[string]interface{}{
				"terms": map[string]interface{}{
					"field":      "status",
					"size":       0,
					"shard_size": 0,
				},
			},
		},
	}

	p.Query = query

	body, err := json.Marshal(p)
	if err != nil {
		fmt.Printf("err :%s \n", err)
		return 0, nil
	}

	fmt.Println(string(body))

	resp, err := Request(OterPlatformURL, "POST", string(body))
	if err != nil {
		fmt.Printf("err :%s \n", err)
		return 0, nil
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("err :%s \n", err)
		return 0, nil
	}

	var result map[string]interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		fmt.Printf("err :%s \n", err)
		return 0, nil
	}

	hits := result["hits"].(map[string]interface{})
	total := hits["total"].(float64)
	return total, hits
	//total = (reflect.TypeOf(result["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})["@timestamp"]))
}

func Time(start_time, end_time string) map[string]interface{} {
	st, _ := time.Parse(STANDARDTIEM, start_time)
	et, _ := time.ParseInLocation(STANDARDTIEM, end_time, time.UTC)
	must3 := map[string]interface{}{
		"range": map[string]interface{}{
			"@timestamp": map[string]string{
				"gte": st.Format(FORMATTIME),
				"lt":  et.Format(FORMATTIME),
			},
		},
	}
	return must3
}

func TestM(t *testing.T) {
	must := Time("2017-01-13 01:20:00", "2017-01-13 01:50:00")
	// tt, d, nex_id := Mem(must, "")
	// fmt.Println("total : ", tt)
	// fmt.Println("nextid :", nex_id)
	// db, _ := json.Marshal(d)
	f0, _ := os.OpenFile("3.data", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0664)

	b := Data(must)
	f0.Write(b)
}

func TestTotal(t *testing.T) {
	f0, _ := os.OpenFile("3.data", os.O_CREATE|os.O_RDWR, 0664)
	var b []byte
	b, err := ioutil.ReadAll(f0)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string][]map[string]int)
	json.Unmarshal(b, &result)
	var total int
	for _, v := range result {
		for _, lv := range v {
			for _, s := range lv {
				total += s
			}
		}
	}
	fmt.Println(total)
}

func Data(must map[string]interface{}) []byte {

	total, _ := Mem(0, 20, must)
	var pages int
	t := int(total)
	if t%20000 == 0 {
		pages = t / 20000
	} else {
		pages = (t / 20000) + 1
	}
	result := make(map[string][]map[string]int)
	res := make(map[string]int)

	fmt.Println("size : ", total)
	for i := 0; i < pages; i++ {
		fmt.Printf("currect :%d ,total : %d \n", i, pages)
		_, hits := Mem(i*20000, 20000, must)
		if hits != nil {
			for _, v := range hits["hits"].([]interface{}) {
				node := v.(map[string]interface{})["_source"].(map[string]interface{})["from_node"].(string)
				result[node] = nil
				_type := v.(map[string]interface{})["_type"].(string)

				if _, ok := res[node+"_"+_type]; ok {
					res[node+"_"+_type] += 1
				} else {
					res[node+"_"+_type] = 1
				}

			}
		}
	}

	for k, v := range res {
		for fk, _ := range result {
			if strings.Contains(k, fk) {
				result[fk] = append(result[fk], map[string]int{k: v})
			}
		}
	}

	fmt.Println(result)
	b, _ := json.Marshal(result)
	return b
}
