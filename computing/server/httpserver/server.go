package httpserver

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/gridprotocol/computing-api/common/utils"
	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/deploy"
	"github.com/gridprotocol/computing-api/computing/gateway"
	"github.com/gridprotocol/computing-api/computing/model"
	"github.com/gridprotocol/computing-api/lib/logc"

	"github.com/gin-gonic/gin"
)

var logger = logc.Logger("http")

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

// handler all requests, including greet and compute
func (hc *handlerCore) handlerAllRequests(c *gin.Context) {
	// handle the greet requst
	if c.Request.URL.Path == "/greet" && c.Request.Method == "GET" {
		hc.handlerGreet(c)
		return
	} else {
		// handle the compute request
		hc.handlerCompute(c)
	}
}

// handler greet request
func (hc *handlerCore) handlerGreet(c *gin.Context) {
	// greet type
	msgType := c.Query("type")
	user := c.Query("user")

	// get cp address from config file
	cp := config.GetConfig().Remote.Wallet

	logger.Debug("user:", user)
	logger.Debug("cp:", cp)

	// for each greet type
	switch msgType {

	// provider confirm and activate an order
	case "confirm":
		err := hc.gw.Confirm(user)
		if err != nil {
			msg := fmt.Sprintf("[Fail] Failed to confirm: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order confirmed and activated"})

	// send cookie to user
	case "cookie":
		ts := c.Query("ts")
		sig := c.Query("sig")

		/*
			//todo: release when sig scripts is ok
			// check status must be activated
			if orderInfo.Status != 2 {
				var status string
				switch orderInfo.Status {
				case 0:
					status = "order not exist"
				case 3:
					status = "order cancelled"
				case 4:
					status = "order completed"
				}
				c.JSON(http.StatusInternalServerError, gin.H{"msg": "[Fail] only activated order can get cookie:" + status})
				return
			}
		*/

		// check order before send cookie
		ok, err := hc.gw.OrderCheck(user, cp)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("[Fail] %s", err.Error())})
			return
		}

		// verify ts and sig
		if len(ts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] missing timestamp in request"})
			return
		}
		if len(sig) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] missing signature in request"})
			return
		}

		// verify signature in type2
		ok = hc.gw.VerifyAccessibility(&model.AuthInfo{Address: user, Sig: sig, Msg: ts})
		if ok {
			// make cookie from addr and expire
			cookie := hc.cm.MakeCookie(user)

			// set cookie into response
			http.SetCookie(c.Writer, cookie)

			// response with cookie content
			c.JSON(http.StatusOK, gin.H{
				"msg":    "[ACK] user authorized",
				"cookie": cookie.String(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] Failed to verify your signature"})
		}

	// deploy with yaml url
	case "deployUrl":
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

		// yaml url
		url := c.Query("url")

		// missing url
		if len(url) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] missing yaml url in request"})
			return
		}

		// parse url into deps and svcs
		deps, svcs, err := deploy.ParseYamlUrl(url)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("parse yaml url failed:%s", err.Error())})
			return
		}

		logger.Debug("deploying app")
		// deploy with remote yaml file
		err = hc.gw.Deploy(deps, svcs, addr)
		if err != nil {
			// clean all failed deployments from k8s
			for _, dep := range deps {
				// todo: what if delete failed
				_ = deploy.CleanDeploy(dep.Name)
			}

			msg := fmt.Sprintf("[Fail] Failed to deploy: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] deploy from url ok"})

	// deploy by id
	case "deployID":
		yamlID := c.Query("id")
		if len(yamlID) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] missing yaml id in request"})
			return
		}

		// inject a cookie into request header, in case the cookie is refused by the client(browser)
		cks := injectCookie(c)

		// check cookies‘ expire and sig
		addr, err := hc.cm.CheckCookie(cks)
		if err != nil {
			msg := fmt.Sprintf("[Fail] invalid cookie: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"err": msg})
			return
		}

		logger.Info("cookie check passed, addr:", addr)

		// if no remote yaml is provided either, response error
		if len(yamlID) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] the request missing yaml id"})
			return
		}

		// get yaml path from id in input
		p, err := utils.GetPathByID(yamlID)
		if err != nil {
			msg := fmt.Sprintf("[Fail] invalid yaml id: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"msg": msg})
			return
		}

		logger.Info("yaml path from id:", p)

		// parse yaml into deps and svcs
		deps, svcs, err := deploy.ParseYamlFile(p)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("parse yaml failed:%s", err.Error())})
			return
		}

		// deploy with local yaml file data
		err = hc.gw.Deploy(deps, svcs, addr)
		if err != nil {
			// clean all deployments from k8s if error happend when deploy
			for _, dep := range deps {
				// todo: what if delete failed
				_ = deploy.CleanDeploy(dep.Name)
			}

			msg := fmt.Sprintf("[Fail] Failed to deploy: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		logger.Debug("app name:", deps[0].Name)
		// set the app name in order
		err = hc.gw.SetApp(user, deps[0].Name)
		if err != nil {
			// clean all deployments from k8s if error happend when deploy
			for _, dep := range deps {
				// todo: what if delete failed
				_ = deploy.CleanDeploy(dep.Name)
			}

			msg := fmt.Sprintf("[Fail] Failed to set app: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] deploy ok"})

	// order renew
	case "renew":
		sk := c.Query("sk")
		dur := c.Query("dur")
		pay := c.Query("pay")

		logger.Debug("user:", user)
		logger.Debug("sk:", sk)

		// get order info with params
		orderInfo, err := hc.gw.GetOrder(user, cp)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] get order info from contract failed: " + err.Error()})
			return
		}
		logger.Debug("order info:", orderInfo)

		// check order status
		if orderInfo.Status != 2 {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "[Error] only activated order can be renewed"})
			return
		}

		logger.Debug("renewing order")

		// renew order
		err = hc.gw.Renew(user, sk, dur, pay)
		if err != nil {
			msg := fmt.Sprintf("[Fail] Failed to renew: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order renewed"})

	// reset order to status 1(unactive), called by cp
	case "reset":
		prob := c.Query("prob")
		dur := c.Query("dur")

		logger.Debug("user:", user)
		logger.Debug("cp:", cp)

		logger.Debug("reseting order")

		// reset order
		err := hc.gw.Reset(user, cp, prob, dur)
		if err != nil {
			msg := fmt.Sprintf("[Fail] Failed to reset: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order reset to status 1 (unactive)"})

	// provider settle
	case "settle":
		logger.Debug("user:", user)

		logger.Debug("settling order")

		// cancel order
		err := hc.gw.Settle(user)
		if err != nil {
			msg := fmt.Sprintf("[Fail] Failed to settle: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order settle ok"})

	// illegal type
	default:
		c.JSON(http.StatusBadRequest, gin.H{"err": "unsupported msg type"})
	}
}

// for all other requests, forward them to a proxy, and return the response from the proxy to the client
func (hc *handlerCore) handlerCompute(c *gin.Context) {
	// inject a cookie into request header, in case the cookie is refused by the client(browser)
	cks := injectCookie(c)
	logger.Info("cookies:", cks)

	// check cookie's expire and sig, return the user addr in cookie name
	addr, err := hc.cm.CheckCookie(cks)
	if err != nil {
		msg := fmt.Sprintf("cookie check failed: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"msg": msg})
		return
	}

	// get order info and do expire check for it
	// get cp address from config file
	cp := config.GetConfig().Remote.Wallet
	// get order info with params
	orderInfo, err := hc.gw.GetOrder(addr, cp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] get order info from contract failed: " + err.Error()})
		return
	}
	logger.Debug("order info:", orderInfo)
	// order expire check
	ok, err := hc.gw.ExpireCheck(*orderInfo)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] the order expire check failed: " + err.Error()})
		return
	}
	logger.Debug("expire check ok")

	// query entrance url(service endpoint) stored in DB with address
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

		// replace host info in the request with entrance
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

// for the cross domain access
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
