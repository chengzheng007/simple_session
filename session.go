package simple_session

import (
	"encoding/hex"
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"simple_session/store"
	"time"
)

type Config struct {
	SidPrefix       string `json:"sidPrefix"` // sid ptrfix in redis key
	CookieName      string `json:"cookieName"`
	EnableSetCookie bool   `json:"enableSetCookie,omitempty"`
	Gclifetime      int64  `json:"gclifetime"`
	Maxlifetime     int64  `json:"maxLifetime"`
	Secure          bool   `json:"secure"`
	CookieLifeTime  int    `json:"cookieLifeTime"`
	ConnConfig      string `json:"ConnConfig"`
	Domain          string `json:"domain"`
	SessionIDLength int64  `json:"sessionIDLength"`
}

var (
	cookieCfg *Config
)

func Init(cfg Config) error {
	if len(cfg.CookieName) == 0 {
		return errors.New("Invalid CookieName")
	}
	if cfg.Maxlifetime <= 0 {
		cfg.Maxlifetime = cfg.Gclifetime
	}
	if err := store.InitPool(cfg.Maxlifetime, cfg.ConnConfig); err != nil {
		return err
	}

	if cfg.SessionIDLength <= 0 {
		cfg.SessionIDLength = 16
	}
	cookieCfg = &cfg
	return nil
}

func SessionStart(w http.ResponseWriter, r *http.Request) (*store.Store, error) {
	sid, err := getSid(r)
	if err != nil {
		return nil, err
	}

	if sid != "" && store.SessionExist(sid) {
		return store.SessionRead(sid)
	}

	// generate a new session id
	sid = genSid()
	stobj, err := store.SessionRead(sid)
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:     cookieCfg.CookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		Domain:   cookieCfg.Domain,
		Expires:  time.Now().Add(time.Duration(cookieCfg.CookieLifeTime) * time.Second),
		MaxAge:   cookieCfg.CookieLifeTime,
		Secure:   cookieCfg.Secure,
		HttpOnly: false,
	}

	if cookieCfg.EnableSetCookie {
		w.Header().Set("P3P", "IDC DSP COR ADM DEVi TAIi PSA PSD IVAi IVDi CONi HIS OUR IND CNT")
		http.SetCookie(w, cookie)
	}

	r.AddCookie(cookie)

	return stobj, nil
}

// get sid from browser
func getSid(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieCfg.CookieName)

	if err != nil || cookie.Value == "" || cookie.MaxAge < 0 {
		if err = r.ParseForm(); err != nil {
			return "", err
		}
		// make sure it is in GET or POST
		sid := r.FormValue(cookieCfg.CookieName)
		return sid, nil
	}
	return url.QueryEscape(cookie.Value), nil
}

// gen sid
func genSid() string {
	rand.Seed(time.Now().UnixNano())
	sids := make([]byte, cookieCfg.SessionIDLength)
	for i := 0; i < len(sids); i++ {
		sids[i] = byte(rand.Intn(256))
	}
	return cookieCfg.SidPrefix + hex.EncodeToString(sids)
}

func SessionDestroy(wr http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieCfg.CookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	store.SessionDestroy(sid)
	if cookieCfg.EnableSetCookie {
		wr.Header().Set("P3P", "IDC DSP COR ADM DEVi TAIi PSA PSD IVAi IVDi CONi HIS OUR IND CNT")
		cookie = &http.Cookie{
			Name: cookieCfg.CookieName,
			Path: "/",
			// HttpOnly: true,
			Expires: time.Now().Local().Add(-600 * time.Second),
			MaxAge:  -1,
		}
		http.SetCookie(wr, cookie)
	}
}
