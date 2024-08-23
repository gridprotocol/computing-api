package httpserver

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/lib/auth"
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
		expire:  time.Duration(time.Duration(config.GetConfig().Http.CookieExpire) * time.Second),
	}
}

// make a cookie from user addr and expire ts, sign with signKey
func (cm *cookieManager) MakeCookie(addr string) *http.Cookie {
	// calc expire time
	expire := time.Now().Add(cm.expire)
	// time to string
	ts := strconv.FormatInt(expire.Unix(), 10)

	// sign msg with sign key
	sig, _ := auth.SignToken(addr+ts, cm.signKey)

	// make the cookie with the sig and expire
	cookie := &http.Cookie{
		Name:    tokenPrefix + addr,
		Value:   ts + "_" + sig,
		Expires: expire,
	}

	return cookie
}

// find a valid cookie from all cookies in the request
func (cm *cookieManager) FindCookie(cks []*http.Cookie) (string, error) {
	// search for a valid cookie
	for _, ck := range cks {
		// check name format
		if !strings.HasPrefix(ck.Name, tokenPrefix) {
			continue
		}
		// get the user address from the cookie name
		user := ck.Name[len(tokenPrefix):]
		// get the expire ts and cookie signature from the cookie value
		parts := strings.SplitN(ck.Value, "_", 2)

		// check value format
		if len(parts) != 2 {
			return "", fmt.Errorf("the cookie's value format is invalid")
		}

		// get the expire ts and sig of token
		ts, sig := parts[0], parts[1]
		// transfer str into int
		tsInt, err := strconv.Atoi(ts)
		if err != nil {
			return "", err
		}

		// check the sig in the cookie with the sign key in config
		err = auth.VerifyToken(user+ts, sig, cm.signKey)
		if err != nil {
			return "", err
		}

		// check cookie expiration
		if !time.Now().Before(time.Unix(int64(tsInt), 0)) {
			return "", fmt.Errorf("the cookie's expire time is end")
		}

		// if all check passed for this cookie, return the address
		return user, nil
	}

	return "", fmt.Errorf("no valid cookie found")
}
