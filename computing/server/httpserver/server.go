package httpserver

import (
	"computing-api/computing/gateway"
	"computing-api/computing/model"
	"computing-api/lib/logs"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/gin-gonic/gin"
)

var logger = logs.Logger("http")

type handlerCore struct {
	gw  gateway.ComputingGatewayAPI
	rpp sync.Pool // reverse proxy pool
	cm  *cookieManager
}

func NewServer(addr string, gw gateway.ComputingGatewayAPI) *http.Server {
	logger.Info("Start server")
	gin.SetMode(gin.ReleaseMode)
	route := registerAllRoute(gw)
	server := &http.Server{
		Addr:    addr,
		Handler: route,
	}
	logger.Info("Set route ok")
	return server
}

func registerAllRoute(gw gateway.ComputingGatewayAPI) *gin.Engine {
	route := gin.Default()
	route.Use(cors())
	hc := handlerCore{
		gw: gw,
		cm: newCookieManager(),
		rpp: sync.Pool{
			New: func() any {
				return &httputil.ReverseProxy{}
			},
		},
	}
	// route.GET("/greet", hc.handlerGreet)
	route.Any("/*path", hc.handlerProcess)
	return route
}

func (hc *handlerCore) handlerGreet(c *gin.Context) {
	msgType := c.Query("type")
	input := c.Query("input")
	switch msgType {
	case "0": // lease
		if hc.gw.CheckContract(input) {
			c.JSON(http.StatusOK, gin.H{"msg": "[ACK] the lease is acceptable"})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "[Fail] the lease is not acceptable"})
		}
	case "1": // apply for authority
		// TODO: cache check
		if !hc.gw.CheckContract(input) {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] the lease is not acceptable"})
			return
		}
		// check payee (send activate tx if necessary)
		_, user := hc.gw.CheckPayee(input)
		if len(user) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] no user address is provided"})
			return
		}
		if err := hc.gw.Authorize(user, model.Lease{}); err != nil {
			logger.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "[Fail] failed to authorized"})
			return
		}
		// set watcher for the lease (current ver is empty)
		hc.gw.SetWatcher(input)
		c.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("[ACK] %s authorized ok", user)})
	case "2": // Acquire cookie for later access
		addr := c.Query("addr")
		sig := c.Query("sig")
		if hc.gw.VerifyAccessibility(&model.AuthInfo{Address: addr, Sig: sig, Msg: input}) {
			cookie := hc.cm.Set(addr)
			http.SetCookie(c.Writer, cookie)
			c.JSON(http.StatusOK, gin.H{"msg": "[ACK] already authorized"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] Failed to verify your signature"})
		}
	case "3": // deploy
		// verify accessibility
		addr, ok := hc.cm.CheckCookie(c.Request.Cookies())
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"err": "invalid cookie"})
			return
		}
		// local for test
		localfile := c.Query("local")
		local := false
		if len(localfile) != 0 {
			input = localfile
			local = true
		}
		// input is deploy-yaml-file-url
		if len(input) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] empty deployment"})
			return
		}
		err := hc.gw.Deploy(addr, input, local)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "[Fail] Failed to deploy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] deployed ok"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"err": "unsupported msg type"})
	}
}

func (hc *handlerCore) handlerProcess(c *gin.Context) {
	// redirect handler
	if c.Request.URL.Path == "/greet" && c.Request.Method == "GET" {
		hc.handlerGreet(c)
		return
	}

	// verify accessibility
	addr, ok := hc.cm.CheckCookie(c.Request.Cookies())
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"err": "invalid cookie"})
		return
	}

	// forward entrance
	ent, err := hc.gw.GetEntrance(addr)
	if err != nil {
		logger.Error("No Entrance: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] have not deployed before or something went wrong"})
		return
	}
	logger.Info(ent)
	targetURL, err := url.Parse(ent)
	if err != nil {
		logger.Error("Fail to parse url: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"err": "fail to parse entrance"})
		return
	}
	logger.Info(targetURL)

	// forward rules
	director := func(r *http.Request) {
		if len(targetURL.Scheme) != 0 {
			r.URL.Scheme = targetURL.Scheme
		} else {
			r.URL.Scheme = "http"
		}
		r.URL.Host = targetURL.Host
		r.Host = targetURL.Host
	}
	proxy := hc.rpp.Get().(*httputil.ReverseProxy)
	defer hc.rpp.Put(proxy)
	proxy.Director = director

	proxy.ServeHTTP(c.Writer, c.Request)
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
