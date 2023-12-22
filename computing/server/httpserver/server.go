package httpserver

import (
	"computing-api/computing/gateway"
	"computing-api/computing/model"
	"computing-api/lib/logs"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var logger = logs.Logger("http")

type handlerCore struct {
	gw gateway.ComputingGatewayAPI
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
	route.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"route":   "/greet and /process",
			"greet":   "required: type (0-3), input",
			"process": "required: address", // http request (forward), api_key
		})
	})

	hc := handlerCore{gw}
	route.GET("/greet", hc.handlerGreet)
	route.GET("/process", hc.handlerProcess)
	return route
}

func (hc handlerCore) handlerGreet(c *gin.Context) {
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
	case "2": // check authority
		if hc.gw.VerifyAccessibility(input, "", false) {
			c.JSON(http.StatusOK, gin.H{"msg": "[ACK] already authorized"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] Failed to verify your account"})
		}
	case "3": // deploy
		// input is deploy-yaml-file-url
		if len(input) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] empty deployment"})
			return
		}
		addr := c.Query("addr")
		if len(addr) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] empty user address"})
			return
		}
		if !hc.gw.VerifyAccessibility(addr, "", false) {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] user is not authorized"})
			return
		}
		// file path for test
		input = "./tomcat-dm.yaml" // if loacal=true, set this local file
		err := hc.gw.Deploy(addr, input, true)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"msg": "[ACK] deployed failed"})
		}
		c.JSON(http.StatusOK, gin.H{"msg": "[ACK] deployed ok"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"err": "unsupported msg type"})
	}
}

// temporarily ignore api_key
func (hc handlerCore) handlerProcess(c *gin.Context) {
	addr := c.Query("addr")
	if !hc.gw.VerifyAccessibility(addr, "", false) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] user is not authorized"})
		return
	}
	ent, err := hc.gw.GetEntrance(addr)
	if err != nil {
		logger.Error("No Entrance: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] have not deployed before or something went wrong"})
		return
	}
	// temp fixate the request to get /
	in := model.ComputingInput{Request: nil}
	out := model.ComputingOutput{Response: nil}
	err = hc.gw.Compute(ent, &in, &out)
	if err != nil {
		logger.Error("Bad request: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("[Fail] failed to compute: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] compute ok", "response": out.Response})
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
