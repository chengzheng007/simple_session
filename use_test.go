package simple_session

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestSession(t *testing.T) {

	// initialize config
	conf := Config{}
	// cookie name from browser
	conf.CookieName = "mys"
	// session store, redis key prefix
	conf.SidPrefix = "mysession_"

	// redis connect ip:port
	conf.ConnConfig = "127.0.0.1:6379"

	conf.Gclifetime = 86400 * 7
	conf.Maxlifetime = 86400 * 7
	conf.CookieLifeTime = 86400 * 7
	conf.SessionIDLength = 16

	conf.Domain = ".my-website.com"

	conf.EnableSetCookie = true
	conf.Secure = false

	if err := Init(conf); err != nil {
		t.Fatalf("session Init error(%v)", err)
		return
	}

	// ... start serve request
	startHandle()
}

func startHandle() {
	http.HandleFunc("set_session", setSession)
	http.HandleFunc("get_session", getSession)
	http.ListenAndServe("127.0.0.1:9000", nil)
}

// set session to redis
func setSession(w http.ResponseWriter, r *http.Request) {

	res := make(map[string]interface{})
	defer retWriter(w, res)

	store, err := SessionStart(w, r)
	if err != nil {
		res["msg"] = "SessionStart error"
		return
	}
	defer store.Persistence()

	name := "Jack"
	age := 25

	store.Set("name", name)
	store.Set("age", age)

	res["msg"] = "OK"
	return
}

// get session data from redis, return back to remote client
func getSession(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]interface{})
	defer retWriter(w, res)

	store, err := SessionStart(w, r)
	if err != nil {
		res["msg"] = "SessionStart error"
		return
	}

	res["name"] = store.Get("name")
	res["age"] = store.Get("age")
	res["msg"] = "OK"
	return
}

func retWriter(w http.ResponseWriter, res map[string]interface{}) {
	byt, err := json.Marshal(res)
	if err != nil {
		fmt.Printf(" json.Marshal(%#v) error(%v)", res, err)
		return
	}
	if _, err := w.Write(byt); err != nil {
		fmt.Printf("w.Write(%s) error(%v)", byt, err)
	}
}
