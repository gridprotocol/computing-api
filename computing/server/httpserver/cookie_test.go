package httpserver

import (
	"net/http"
	"testing"
	"time"
)

func TestCookieProcess(t *testing.T) {
	var (
		key    = "memo.io"
		expire = time.Hour
		addr   = "0x1234567890123456789012345678901234567890"
	)
	ckm := &cookieManager{
		signKey: []byte(key),
		expire:  expire,
	}

	ck := ckm.MakeCookie(addr)

	// right
	addr2, err := ckm.CheckCookie([]*http.Cookie{ck})
	if err != nil {
		t.Error("fail to check cookie")
	}
	if addr2 != addr {
		t.Error("address extracted from cookie is not matched to the original one")
	}
	t.Log("CheckCookie normal process is ok")

	// expire
	ckm.expire = time.Second
	ck2 := ckm.MakeCookie(addr)
	time.Sleep(time.Second)
	_, err = ckm.CheckCookie([]*http.Cookie{ck2})
	if err != nil {
		t.Error("cookie should be expired")
	} else {
		t.Log("Expire test is ok")
	}

	// bad signature
	ckm.expire = time.Hour
	ck3 := *ck
	ck3.Value = "2" + ck.Value[1:]
	_, err = ckm.CheckCookie([]*http.Cookie{&ck3})
	if err != nil {
		t.Error("should be invalid signature")
	} else {
		t.Log("invalid signature test is ok")
	}
}
