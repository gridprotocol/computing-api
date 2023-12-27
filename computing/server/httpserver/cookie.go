package httpserver

import (
	"computing-api/lib/tool"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	tokenPrefix = "cpuser_"
)

type cookieManager struct {
	// cookie cache
	cc map[string]string
	mu sync.RWMutex
}

func newCookieManager() *cookieManager {
	return &cookieManager{
		cc: make(map[string]string),
	}
}

func (cm *cookieManager) Set(addr string) *http.Cookie {
	token := tool.GenerateRandomString(10)
	cookie := &http.Cookie{
		Name:    tokenPrefix + addr,
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour),
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cc[addr] = token
	return cookie
}

func (cm *cookieManager) get(addr string) (string, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	token, ok := cm.cc[addr]
	return token, ok
}

func (cm *cookieManager) Delete(addr string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.cc, addr)
}

func (cm *cookieManager) CheckCookie(cks []*http.Cookie) (string, bool) {
	for _, ck := range cks {
		if strings.HasPrefix(ck.Name, tokenPrefix) {
			addr := ck.Name[len(tokenPrefix):]
			if validToken, ok := cm.get(addr); ok {
				if validToken == ck.Value {
					return addr, true
				}
			}
			return "", false
		}
	}
	return "", false
}
