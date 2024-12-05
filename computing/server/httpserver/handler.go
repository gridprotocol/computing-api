package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/deploy"
	"github.com/gridprotocol/computing-api/computing/docker"
	"github.com/gridprotocol/computing-api/computing/model"
	"github.com/gridprotocol/computing-api/lib/utils"

	"github.com/gin-gonic/gin"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
func (hc *handlerCore) handlerConfirm(c *gin.Context) {
	user := c.Query("user")

	// call confirm
	err := hc.gw.Confirm(user)
	if err != nil {
		msg := fmt.Sprintf("[Fail] Failed to confirm: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order confirmed and activated"})

}
*/

/*
func (hc *handlerCore) handlerActivate(c *gin.Context) {
	user := c.Query("user")

	// call confirm
	err := hc.gw.Activate(user)
	if err != nil {
		msg := fmt.Sprintf("[Fail] Failed to confirm: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order activated"})

}

func (hc *handlerCore) handlerDeactivate(c *gin.Context) {
	user := c.Query("user")

	// call confirm
	err := hc.gw.Deactivate(user)
	if err != nil {
		msg := fmt.Sprintf("[Fail] Failed to confirm: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order deactivated"})

}
*/

func (hc *handlerCore) handlerCookie(c *gin.Context) {
	// user address
	user := c.Query("user")

	// get cp address from config file
	//cp := config.GetConfig().Remote.Wallet

	ts := c.Query("ts")
	sig := c.Query("sig")

	/*
		// check order before send cookie
		ok, err := hc.gw.OrderCheck(oid64)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("[Fail] %s", err.Error())})
			return
		}
	*/

	if len(ts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] missing timestamp in request"})
		return
	}
	if len(sig) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] missing signature in request"})
		return
	}

	// check auth info, signature and it's expire
	ok := hc.gw.CheckAuthInfo(&model.AuthInfo{Address: user, Sig: sig, Msg: ts})
	// check passed, make and send a cookie
	if ok {
		// make cookie from addr and cookie expire
		cookie := hc.cm.MakeCookie(user)

		logger.Debug("new cookie:", cookie)

		// set cookie
		//http.SetCookie(c.Writer, cookie)
		c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)

		// response with cookie content
		c.JSON(http.StatusOK, gin.H{
			"msg":    "[ACK] user authorized",
			"cookie": cookie.String(),
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] Failed to verify your signature"})
	}
}

func (hc *handlerCore) handlerDeployUrl(c *gin.Context) {
	// inject a cookie into request header, in case the cookie is refused by the client(browser)
	cks := injectCookie(c)

	// find a valid cookie in request, and return the user address in this cookie
	user, err := hc.cm.FindCookie(cks)
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
	err = hc.gw.Deploy(deps, svcs, user)
	if err != nil {
		deploy.Clean(deps)

		msg := fmt.Sprintf("[Fail] Failed to deploy: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] deploy from url ok"})
}

// deploy app by app id
func (hc *handlerCore) handlerDeployID(c *gin.Context) {
	oid := c.Query("oid")
	oid64, _ := utils.StringToUint64(oid)

	yamlID := c.Query("id")
	if len(yamlID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] missing yaml id in request"})
		return
	}

	// inject a cookie into request header, in case the cookie is refused by the client(browser)
	//cks := injectCookie(c)

	// get all cookie in the request
	cks := c.Request.Cookies()

	// try to find a valid cookie
	user, err := hc.cm.FindCookie(cks)
	if err != nil {
		msg := fmt.Sprintf("[Fail] invalid cookie: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"err": msg})
		return
	}

	logger.Info("cookie check passed, addr:", user)

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

	// get cp address from config file
	cp := config.GetConfig().Remote.Wallet
	logger.Info("cp addr:", cp)

	// get order info with params
	orderInfo, err := hc.gw.GetOrder(oid64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] get order info from contract failed: " + err.Error()})
		return
	}
	logger.Debug("node id:", orderInfo.NodeId)

	// set node id for the first deploy
	deps[0].Spec.Template.Spec.NodeSelector["id"] = utils.Uint64ToString(orderInfo.NodeId)

	// deploy deps
	err = hc.gw.Deploy(deps, svcs, user)
	if err != nil {
		deploy.Clean(deps)

		msg := fmt.Sprintf("[Fail] Failed to deploy: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	logger.Debug("app name:", deps[0].Name)
	// set the app name in order
	err = hc.gw.SetApp(oid64, deps[0].Name)
	if err != nil {
		deploy.Clean(deps)

		msg := fmt.Sprintf("[Fail] Failed to set app: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] deploy ok"})
}

// clean deploy
func (hc *handlerCore) handlerClean(c *gin.Context) {
	// inject a cookie into request header, in case the cookie is refused by the client(browser)
	//cks := injectCookie(c)

	oid := c.Query("oid")
	oid64, _ := utils.StringToUint64(oid)

	// get all cookie in the request
	cks := c.Request.Cookies()

	// try to find a valid cookie
	user, err := hc.cm.FindCookie(cks)
	if err != nil {
		msg := fmt.Sprintf("[Fail] invalid cookie: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"err": msg})
		return
	}
	logger.Info("cookie check passed, addr:", user)

	// get cp address from config file
	cp := config.GetConfig().Remote.Wallet
	logger.Info("cp addr:", cp)

	// get order info with params
	orderInfo, err := hc.gw.GetOrder(oid64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] get order info from contract failed: " + err.Error()})
		return
	}
	logger.Debug("order info:", orderInfo)

	// delete app
	deploy.DelDeploy(orderInfo.AppName)

	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] clean ok"})
}

// show current app
func (hc *handlerCore) handlerShow(c *gin.Context) {
	// inject a cookie into request header, in case the cookie is refused by the client(browser)
	//cks := injectCookie(c)

	oid := c.Query("oid")
	oid64, _ := utils.StringToUint64(oid)

	// get all cookie in the request
	cks := c.Request.Cookies()

	// try to find a valid cookie
	user, err := hc.cm.FindCookie(cks)
	if err != nil {
		msg := fmt.Sprintf("[Fail] invalid cookie: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"err": msg})
		return
	}
	logger.Info("cookie check passed, addr:", user)

	// get cp address from config file
	cp := config.GetConfig().Remote.Wallet
	logger.Info("cp addr:", cp)

	// get order info with params
	orderInfo, err := hc.gw.GetOrder(oid64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] get order info from contract failed: " + err.Error()})
		return
	}
	logger.Debug("order info:", orderInfo)

	// get app name from order
	deployName := orderInfo.AppName

	if deployName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] no app deployed for this order"})
		return
	}

	// get k8s service
	k8s := docker.NewK8sService()
	// get current version of deployment
	deploymentsClient := k8s.Clientset.AppsV1().Deployments(corev1.NamespaceDefault)
	result, err := deploymentsClient.Get(context.TODO(), deployName, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] get deployment info failed: " + err.Error()})
		return
	}

	// deployment status
	avail := false
	progress := false

	// get deployment conditions
	conditions := result.Status.Conditions
	for _, c := range conditions {
		// check available to be True
		if c.Type == "Available" && c.Status == "True" {
			avail = true
		}
		// progress status
		if c.Type == "Progressing" && c.Status == "True" {
			progress = true
		}
	}

	c.JSON(http.StatusOK, gin.H{"deployment": deployName, "available": avail, "progressing": progress})
}

func (hc *handlerCore) handlerExtend(c *gin.Context) {
	// order id
	oid := c.Query("oid")
	oid64, _ := utils.StringToUint64(oid)

	user := c.Query("user")
	// get cp address from config file
	//cp := config.GetConfig().Remote.Wallet

	sk := c.Query("sk")
	dur := c.Query("dur")

	logger.Debug("user:", user)
	logger.Debug("sk:", sk)

	// get order info with params
	orderInfo, err := hc.gw.GetOrder(oid64)
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
	err = hc.gw.Extend(sk, oid64, dur)
	if err != nil {
		msg := fmt.Sprintf("[Fail] Failed to renew: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order renewed"})
}

func (hc *handlerCore) handlerReset(c *gin.Context) {
	// order id
	oid := c.Query("oid")
	id64, _ := utils.StringToUint64(oid)

	// user addr
	user := c.Query("user")
	// get cp address from config file
	cp := config.GetConfig().Remote.Wallet

	prob := c.Query("prob")
	dur := c.Query("dur")

	logger.Debug("user:", user)
	logger.Debug("cp:", cp)

	logger.Debug("reseting order")

	// reset order
	err := hc.gw.Reset(id64, prob, dur)
	if err != nil {
		msg := fmt.Sprintf("[Fail] Failed to reset: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order reset to status 2 (active)"})
}

func (hc *handlerCore) handlerSettle(c *gin.Context) {
	// order id
	oid := c.Query("oid")
	oid64, _ := utils.StringToUint64(oid)

	// user addr
	user := c.Query("user")

	logger.Debug("user:", user)

	logger.Debug("settling order")

	// cancel order
	err := hc.gw.Settle(oid64)
	if err != nil {
		msg := fmt.Sprintf("[Fail] Failed to settle: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "[ACK] order settle ok"})
}

// for all other requests, forward them to a proxy, and return the response from the proxy to the client
func (hc *handlerCore) handlerCompute(c *gin.Context) {
	// inject a cookie into request header, in case the cookie is refused by the client(browser)
	//cks := injectCookie(c)

	// order id
	id := c.Query("id")
	id64, _ := utils.StringToUint64(id)

	// get all cookies in the request
	cks := c.Request.Cookies()
	logger.Info("cookies:", cks)

	// find a valid cookie
	user, err := hc.cm.FindCookie(cks)
	if err != nil {
		msg := fmt.Sprintf("cookie check failed: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"msg": msg})
		return
	}

	// get order info and do expire check for it
	// get cp address from config file
	cp := config.GetConfig().Remote.Wallet

	logger.Info("user in cookie: ", user)
	logger.Info("cp: ", cp)

	// get order info with params
	orderInfo, err := hc.gw.GetOrder(id64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] get order info from contract failed: " + err.Error()})
		return
	}
	logger.Debug("order info:", orderInfo)

	// check status must be activated
	if orderInfo.Status != 2 {
		var status string
		switch orderInfo.Status {
		case 0:
			status = "order not exist"
		case 1:
			status = "order unactive"
		case 3:
			status = "order cancelled"
		case 4:
			status = "order completed"
		}

		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] order not active: " + status})
		return
	}

	// order expire check
	ok, err := hc.gw.ExpireCheck(*orderInfo)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] the order expire check failed: " + err.Error()})
		return
	}
	logger.Debug("expire check ok")

	// query entrance url(service endpoint) stored in DB with address
	ent, err := hc.gw.GetEntrance(user)
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

type Model struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
	GPU  string `json:"gpu"`
	MEM  string `json:"mem"`
	DISK string `json:"disk"`
}

// get the model list supported
func (hc *handlerCore) handlerModelList(c *gin.Context) {
	logger.Debug("read model list")

	// load model info
	b, err := utils.LoadModel("./list.json")
	if err != nil {
		msg := fmt.Sprintf("[Fail] load model list failed: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	// 将JSON数据解码到一个Person切片
	var models []Model
	err = json.Unmarshal([]byte(b), &models)
	if err != nil {
		msg := fmt.Sprintf("[Fail] Error unmarshalling JSON: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
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

// // handler all requests, including greet and compute
// func (hc *handlerCore) handlerAllRequests(c *gin.Context) {
// 	// handle the greet requst
// 	if c.Request.URL.Path == "/greet" && c.Request.Method == "GET" {
// 		hc.handlerGreet(c)
// 		return
// 	} else {
// 		// handle the compute request
// 		hc.handlerCompute(c)
// 	}
// }
