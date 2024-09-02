package httpserver

import (
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/gridprotocol/computing-api/computing/gateway"
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
	//r.Any("/*path", hc.handlerAllRequests)
	//r.GET("/greet/confirm", hc.handlerConfirm)
	//r.GET("/greet/activate", hc.handlerActivate)
	//r.GET("/greet/deactivate", hc.handlerDeactivate)
	r.GET("/greet/cookie", hc.handlerCookie)
	r.GET("/greet/deployurl", hc.handlerDeployUrl)
	r.GET("/greet/deployid", hc.handlerDeployID)
	r.GET("/greet/renew", hc.handlerRenew)
	r.GET("/greet/reset", hc.handlerReset)
	r.GET("/greet/settle", hc.handlerSettle)
	r.GET("/greet/clean", hc.handlerClean)
	r.GET("/greet/show", hc.handlerShow)
	r.GET("/greet/modellist", hc.handlerModelList)

	r.Any("/", hc.handlerCompute)
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
