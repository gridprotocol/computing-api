package httpserver

import (
	"computing-api/common/utils"
	"computing-api/computing/gateway"
	"computing-api/computing/model"
	"computing-api/lib/logs"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var logger = logs.Logger("http")

type handlerCore struct {
	gw  gateway.ComputingGatewayAPI
	rpp sync.Pool // reverse proxy pool
	cm  *cookieManager
}

// make a new server with a router registered all routes
func NewServer(addr string, gw gateway.ComputingGatewayAPI) *http.Server {
	logger.Info("Starting server")

	// gin mode
	gin.SetMode(gin.ReleaseMode)

	// make a router
	r := gin.Default()

	// register all routes for the new router
	registerAllRoutes(gw, r)

	// new server object with router
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	logger.Info("create new router ok")

	return server
}

// register all routes
func registerAllRoutes(gw gateway.ComputingGatewayAPI, r *gin.Engine) {
	// use middleware for
	r.Use(cors())

	// new hc object with gw
	hc := handlerCore{
		gw: gw,
		cm: newCookieManager(),
		rpp: sync.Pool{
			New: func() any {
				return &httputil.ReverseProxy{}
			},
		},
	}

	// register routes
	//r.GET("/greet", hc.handlerGreet)
	r.Any("/*path", hc.handlerAllRequests)
}

// handler greet request
func (hc *handlerCore) handlerGreet(c *gin.Context) {
	// greet type
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
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] missing address in request"})
			return
		}
		if err := hc.gw.Authorize(user, model.Lease{}); err != nil {
			logger.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "[Fail] authorization failed"})
			return
		}
		// set watcher for the lease (current ver is empty)
		hc.gw.SetWatcher(input)
		c.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("[ACK] %s authorized ok", user)})

	case "2": // Acquire cookie for later access
		addr := c.Query("addr")
		sig := c.Query("sig")
		// verify signature in type2
		ok := hc.gw.VerifyAccessibility(&model.AuthInfo{Address: addr, Sig: sig, Msg: input})
		if ok {
			// make cookie from addr and expire
			cookie := hc.cm.MakeCookie(addr)
			// set cookie into response
			http.SetCookie(c.Writer, cookie)
			// response with cookie
			c.JSON(http.StatusOK, gin.H{
				"msg":    "[ACK] already authorized",
				"cookie": cookie.String(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] Failed to verify your signature"})
		}

	// deploy with input string
	case "3":
		// inject a cookie into request header, in case the cookie is refused by the client(browser)
		cks := injectCookie(c)

		// check the cookie's expire and sig
		addr, err := hc.cm.CheckCookie(cks)
		if err != nil {
			msg := fmt.Sprintf("invalid cookie: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"err": msg})
			return
		}

		logger.Info("cookie check passed")

		// deploy with local filepath or remote file url
		var isLocal = false

		// yaml from local filepath
		localfile := c.Query("local")
		// if local is set
		if len(localfile) != 0 {
			input = localfile
			isLocal = true
		}

		// if no remote yaml is provided either, response error
		if len(input) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] missing yaml resource"})
			return
		}

		// deploy with local or remote yaml file
		err = hc.gw.Deploy(addr, input, isLocal)
		if err != nil {
			msg := fmt.Sprintf("[Fail] Failed to deploy: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] deployed ok"})

	case "4": // deploy by id of local list
		// inject a cookie into request header, in case the cookie is refused by the client(browser)
		cks := injectCookie(c)

		// check the cookie's expire and sig
		addr, err := hc.cm.CheckCookie(cks)
		if err != nil {
			msg := fmt.Sprintf("[Fail] invalid cookie: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"err": msg})
			return
		}

		logger.Info("cookie check passed")

		// if no remote yaml is provided either, response error
		if len(input) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] the request missing yaml id"})
			return
		}

		// get yaml path from id in input
		p, err := utils.GetPathByID(input)
		if err != nil {
			msg := fmt.Sprintf("[Fail] invalid yaml id: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"msg": msg})
			return
		}

		logger.Info("yaml path from id:", p)

		// deploy with local yaml path
		err = hc.gw.Deploy(addr, p, true)
		if err != nil {
			msg := fmt.Sprintf("[Fail] Failed to deploy: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] deployed ok"})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"err": "unsupported msg type"})
	}
}

// handler all requests
func (hc *handlerCore) handlerAllRequests(c *gin.Context) {
	// handle the greet requst preferentially
	if c.Request.URL.Path == "/greet" && c.Request.Method == "GET" {
		hc.handlerGreet(c)
		return
	}

	// for all other requests, forward them to a proxy, and return the response from the proxy to the client

	// inject a cookie into request header, in case the cookie is refused by the client(browser)
	cks := injectCookie(c)
	logger.Info("cookies:", cks)

	// check cookie's expire and sig, return the addr in cookie name
	addr, err := hc.cm.CheckCookie(cks)
	if err != nil {
		msg := fmt.Sprintf("cookie check failed: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"msg": msg})
		return
	}

	// query entrance url stored in DB with address
	ent, err := hc.gw.GetEntrance(addr)
	if err != nil {
		logger.Error("No Entrance: ", err)
		msg := fmt.Sprintf("[Fail] have not deployed before or something went wrong: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"msg": msg})
		return
	}
	logger.Info("entrance:", ent)

	// parse the entrance url into an URL struct
	targetURL, err := url.Parse(ent)
	if err != nil {
		logger.Error("Fail to parse url: ", err)
		msg := fmt.Sprintf("fail to parse entrance: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"err": msg})
		return
	}

	// forward rule func
	director := func(r *http.Request) {
		// scheme provided in the target url
		if len(targetURL.Scheme) != 0 {
			r.URL.Scheme = targetURL.Scheme
		} else {
			// if no scheme given, default to http
			r.URL.Scheme = "http"
		}

		// replace host info
		r.URL.Host = targetURL.Host
		r.Host = targetURL.Host
	}

	// get a proxy from pool
	proxy := hc.rpp.Get().(*httputil.ReverseProxy)
	defer hc.rpp.Put(proxy)

	// set the director func for the proxy
	proxy.Director = director

	// redirect requests to proxy, and get response from it
	proxy.ServeHTTP(c.Writer, c.Request)
}

// make a cookie from the auth data in the request header, and inject it into the request header, return all cookies
func injectCookie(c *gin.Context) []*http.Cookie {
	var cks []*http.Cookie
	tokenStr := c.GetHeader("Authorization")
	if len(tokenStr) != 0 {
		parts := strings.SplitN(tokenStr, " ", 2)
		// cookie must begin with Bearer
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return cks
		}
		// add cookie into request header
		c.Request.Header.Add("Cookie", parts[1])
	}

	// get all cookies in the request header
	cks = c.Request.Cookies()
	return cks
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, Authorization, Token")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
