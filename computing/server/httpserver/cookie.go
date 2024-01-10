package httpserver

import (
	"computing-api/computing/config"
	"computing-api/lib/auth"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	tokenPrefix = "cpuser_"
)

type cookieManager struct {
	signKey []byte
	expire  time.Duration
}

func newCookieManager() *cookieManager {
	return &cookieManager{
		signKey: []byte(config.GetConfig().Http.HSKey),
		expire:  time.Duration(time.Duration(config.GetConfig().Http.Expire) * time.Second),
	}
}

func (cm *cookieManager) Set(addr string) *http.Cookie {
	expire := time.Now().Add(cm.expire)
	ts := strconv.FormatInt(expire.Unix(), 10)
	sig, _ := auth.SignToken(addr+ts, cm.signKey)
	cookie := &http.Cookie{
		Name:    tokenPrefix + addr,
		Value:   ts + "_" + sig,
		Expires: expire,
	}
	return cookie
}

func (cm *cookieManager) CheckCookie(cks []*http.Cookie) (string, bool) {
	for _, ck := range cks {
		if strings.HasPrefix(ck.Name, tokenPrefix) {
			addr := ck.Name[len(tokenPrefix):]
			parts := strings.Split(ck.Value, "_")
			if len(parts) == 2 {
				ts, sig := parts[0], parts[1]
				expire, err := strconv.Atoi(ts)
				if err == nil {
					if time.Now().Before(time.Unix(int64(expire), 0)) {
						if err = auth.VerifyToken(addr+ts, sig, cm.signKey); err == nil {
							return addr, true
						}
					}
				}
			}
			return "", false
		}
	}
	return "", false
}
