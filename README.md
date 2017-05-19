# simple_session
simple session read and write session data base on redis, it refer to beego, but it was very simple and light


# example 
* init configuration variable
```json
	import (
		"github.com/chengzheng007/simple_session"
		"fmt"
	)
	
	/* set your own session config */
	config := simple_session.Config{
		
	}
	err := simple_session.Init(config)
	if err != nil {
		fmt.Println("init session config failed")
	}
```

* start session, set or get variable
```json
	// import necessary package

	// w: http.ResponseWriter, r: *http.Request
	sess, err := simple_session.SessionStart(w, r) 
	if err != nil {
		fmt.Println("session start failed")
	}
	
	// set variable 
	sess.Set("username", 100)
	
	// persistence variable to redis(store in redis really)
	if err := sess.Persistence(); err != nil {
		fmt.Println("sess.Persistence() failed")
	}
	
	// get variable
	val := sess.Get("username")
	// here may need type assertion, because Get always return interface{}
	valI64, _ := val.(int64)

```
