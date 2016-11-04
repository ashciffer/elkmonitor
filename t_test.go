package main

import (
	"encoding/json"
	"io/ioutil"
	//"reflect"
	"strconv"
	"testing"
	"time"
)

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
	must1 := map[string]interface{}{
	 	"match": map[string]interface{}{
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
			"@timestamp": map[string]string{
				"gte": "2016-11-01T02:04:57.780Z",
				"lt":  "2016-11-02T02:09:57.780Z",
			},
		},
	}
	//must4 := map[string]interface{}{
	//	"match": map[string]interface{}{
	//		"msg_id": "581808F7C0A817297E517433C1D1DA9B",
	//	},
	//}

	query := map[string]interface{}{
		"sort": map[string]interface{}{
			"time": map[string]string{
				"order": "desc",
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					must1,
					must2,
					must3,
					//must4,
				},
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
		t.Fatal(err)
	}

	t.Log(string(body))

	resp, err := Request(TaobaoRpcURL, "POST", string(body))
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
	//t.Log()

	//t.Log(reflect.TypeOf(result["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})["@timestamp"]))
}
