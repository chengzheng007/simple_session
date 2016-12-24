package session

import (
	"encoding/hex"
	"errors"
	"git.liebaopay.com/cm_life/bank/dao/session/store"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	SidPrefix       string `json:"sidPrefix"` // sid前缀
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

	// 新会话
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

// 获取浏览器带过来的sid
func getSid(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieCfg.CookieName)

	if err != nil || cookie.Value == "" || cookie.MaxAge < 0 {
		if err = r.ParseForm(); err != nil {
			return "", err
		}
		// 是否在get/post参数里
		sid := r.FormValue(cookieCfg.CookieName)
		return sid, nil
	}
	return url.QueryEscape(cookie.Value), nil
}

// 生成sid
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
