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

// make cookie from addr and expire
func (cm *cookieManager) MakeCookie(addr string) *http.Cookie {
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

// check cookie's expire and verify sig of token
func (cm *cookieManager) CheckCookie(cks []*http.Cookie) (string, bool) {
	for _, ck := range cks {
		if strings.HasPrefix(ck.Name, tokenPrefix) {
			addr := ck.Name[len(tokenPrefix):]
			parts := strings.SplitN(ck.Value, "_", 2)
			if len(parts) == 2 {
				// expire ts and sig of token
				ts, sig := parts[0], parts[1]
				expire, err := strconv.Atoi(ts)
				if err == nil {
					if time.Now().Before(time.Unix(int64(expire), 0)) {
						// verify the sig of token
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
