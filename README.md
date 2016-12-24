# simple_session
simple session base on redis storage, refer to beego, just for simple useage 


# example 
* init configuration variable
```json
	import (
		"github.com/chengzheng007/simple_session"
		"fmt"
	)
	
	config := simple_session.Config{
		/* ... your own session config */
	}
	err := simple_session.Init(config)
	if err != nil {
		fmt.Println("init session config failed")
		return 
	}
```

* start session, set or get variable
```json
	// import necessary package

	// w: http.ResponseWriter, r: *http.Request
	sess, err := simple_session.SessionStart(w, r) 
	if err != nil {
		fmt.Println("session start failed")
		return
	}
	
	// set variable 
	sess.Set("username", 100)
	
	// persistence variable to redis(store in redis really)
	if err := sess.Persistence(); err != nil {
		log.Fatal("sess.Persistence() failed")
		return
	}
	
	// get variable
	val := sess.Get("username")
	// here may need type assertion, because Get always return interface{}
	valI64, _ := val.(int64)

```