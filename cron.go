package main

import (
	"log"
	"time"

	"git.ishopex.cn/matrix/gatling/lib"
	"gopkg.in/mgo.v2/bson"
)

func GetCronTime(_type string) (string, error) {
	var (
		result map[string]interface{}
	)
	session := Mgo.Clone()

	defer func() {
		session.Close()
	}()

	query := bson.M{"type": _type}
	col := session.DB("elk").C("poll_paltform")
	err := col.Find(query).One(&result)
	if err != nil {
		return "", err
	}

	return lib.GetString(result, "last_do_time", ""), nil
}

//设置初始last_do_time 当前时间 - 15 分钟
func SetCronTime(_type string) {
	var (
		result = make(map[string]interface{})
	)

	session := Mgo.Clone()

	defer func() {
		session.Close()
	}()

	result["modifytime"] = time.Now().Unix()
	result["cron_time"] = 5
	result["last_do_time"] = time.Unix(time.Now().UTC().Unix()-15*60, 0).UTC().Format(FORMATTIME)
	result["step"] = 15
	result["type"] = _type

	col := session.DB("elk").C("poll_paltform")
	err := col.Insert(result)
	if err != nil {
		log.Printf("[Error]初始last_do_time插入失败:", err)
	}

}

func ModiyLastDoTime(_type, t string) error {
	selector := bson.M{"type": _type}

	session := Mgo.Clone()
	defer func() {
		session.Close()
	}()

	change := bson.M{"$set": bson.M{"last_do_time": t, "modifytime": time.Now().Unix()}}
	col := session.DB("elk").C("poll_paltform")
	err := col.Update(selector, change)
	if err != nil {
		return err
	}
	return nil
}
