package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grid/contracts/eth"
	"github.com/grid/contracts/eth/contracts"
	com "github.com/gridprotocol/computing-api/common"
	"github.com/gridprotocol/computing-api/common/version"
	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/docker"
	"github.com/gridprotocol/computing-api/computing/gateway"
	"github.com/gridprotocol/computing-api/computing/gateway/remote"
	"github.com/gridprotocol/computing-api/computing/server/httpserver"
	"github.com/gridprotocol/computing-api/keystore"
	"github.com/gridprotocol/computing-api/lib/logc"
	"github.com/gridprotocol/computing-api/lib/utils"
	"github.com/gridprotocol/computing-api/prover"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	logger = logc.Logger("cmd")
	// quit chan
	quit = make(chan os.Signal, 1)

	// user db records
	//userDB kv.Database
)

var DaemonCmd = &cli.Command{
	Name:  "daemon",
	Usage: "platform daemon",
	Subcommands: []*cli.Command{
		runCmd,
		stopCmd,
	},
}

// run daemon
var runCmd = &cli.Command{
	Name:  "run",
	Usage: "run server",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "deploy or direct forward",
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "chain",
			Aliases: []string{"c"},
			Usage:   "chain to interactivate, local: use local test chain, sepo: use sepo test chain, dev:devvChain, test:testChain",
			Value:   "local",
		},
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"pw"},
			Usage:   "password of current wallet",
			Value:   "computing",
		},
	},
	Action: func(ctx *cli.Context) error {
		test := ctx.Bool("test")
		chain := ctx.String("chain")
		pw := ctx.String("pw")

		// get wallet and sk from keystore
		repo := keystore.Repo
		wallet := config.GetConfig().Remote.Wallet
		ki, err := repo.Get(wallet, pw)
		if err != nil {
			fmt.Println("get key info from wallet failed: ", err.Error())
			return err
		}

		// save all info into common
		com.Password = pw
		com.CP = wallet
		com.SK = ki.SK()

		validator_url := config.GetConfig().Validator.Url
		platform_url := config.GetConfig().Platform.Url

		// check version
		if version.CheckVersion() {
			os.Exit(0)
		}
		log.Println("Current Version:", version.CurrentVersion())

		// new provder and start
		logger.Info("starting prover")
		prover, err := prover.NewGRIDProver(chain, validator_url, ki.SK(), 1)
		if err != nil {
			log.Fatalf("new light node prover: %s\n", err)
		}
		go prover.Start(context.Background())

		// check node online
		go CheckOnline(platform_url, wallet)
		go CheckOrders(platform_url, wallet)

		// chain select for remote gw
		var chain_endpoint string
		switch chain {
		case "local":
			chain_endpoint = eth.Ganache

			// load all addresses from json
			logger.Debug("load addresses")
			// loading contracts
			l := contracts.Local{}
			l.Load()
			logger.Debugf("%+v\n", l)

			if l.Market == "" || l.Access == "" || l.Credit == "" || l.Registry == "" {
				logger.Debug("all contract addresses must exist in json file")
			}
			// save address
			remote.MarketAddr = common.HexToAddress(l.Market)
			remote.AccessAddr = common.HexToAddress(l.Access)
			remote.CreditAddr = common.HexToAddress(l.Credit)
			remote.RegistryAddr = common.HexToAddress(l.Registry)

		case "sepo":
			chain_endpoint = eth.Sepolia

			// load all addresses from json
			logger.Debug("load addresses")
			// loading contracts
			s := contracts.Sepo{}
			s.Load()
			logger.Debugf("%+v\n", s)

			if s.Market == "" || s.Access == "" || s.Credit == "" || s.Registry == "" {
				logger.Debug("all contract addresses must exist in json file")
			}
			// save address
			remote.MarketAddr = common.HexToAddress(s.Market)
			remote.AccessAddr = common.HexToAddress(s.Access)
			remote.CreditAddr = common.HexToAddress(s.Credit)
			remote.RegistryAddr = common.HexToAddress(s.Registry)

		case "dev":
			chain_endpoint = eth.DevChain

			// load all addresses from json
			logger.Debug("load addresses")
			// loading contracts
			d := contracts.Dev{}
			d.Load("dev")
			logger.Debugf("%+v\n", d)

			if d.Market == "" || d.Access == "" || d.Credit == "" || d.Registry == "" {
				logger.Debug("all contract addresses must exist in json file")
			}
			// save address
			remote.MarketAddr = common.HexToAddress(d.Market)
			remote.AccessAddr = common.HexToAddress(d.Access)
			remote.CreditAddr = common.HexToAddress(d.Credit)
			remote.RegistryAddr = common.HexToAddress(d.Registry)

		case "product":
			chain_endpoint = eth.DevChain

			// load all addresses from json
			logger.Debug("load addresses")
			// loading contracts
			d := contracts.Dev{}
			d.Load("product")
			logger.Debugf("%+v\n", d)

			if d.Market == "" || d.Access == "" || d.Credit == "" || d.Registry == "" {
				logger.Debug("all contract addresses must exist in json file")
			}

			// save address
			remote.MarketAddr = common.HexToAddress(d.Market)
			remote.AccessAddr = common.HexToAddress(d.Access)
			remote.CreditAddr = common.HexToAddress(d.Credit)
			remote.RegistryAddr = common.HexToAddress(d.Registry)

		case "test":
			chain_endpoint = eth.TestChain

			// load all addresses from json
			logger.Debug("load addresses")
			// loading contracts
			t := contracts.Test{}
			t.Load()
			logger.Debugf("%+v\n", t)

			if t.Market == "" || t.Access == "" || t.Credit == "" || t.Registry == "" {
				logger.Debug("all contract addresses must exist in json file")
			}
			// save address
			remote.MarketAddr = common.HexToAddress(t.Market)
			remote.AccessAddr = common.HexToAddress(t.Access)
			remote.CreditAddr = common.HexToAddress(t.Credit)
			remote.RegistryAddr = common.HexToAddress(t.Registry)

		default:
			log.Fatal("unsupport chain")
		}

		logger.Debug("platform url:", platform_url)

		// make a gw object
		gw := gateway.NewComputingGateway(chain_endpoint, platform_url, wallet, test)
		// close db
		defer gw.Close()

		logger.Debug("listen address: ", config.GetConfig().Http.Listen)

		// make an httpserver with listen addr and gw object
		svr := httpserver.NewServer(config.GetConfig().Http.Listen, gw)
		// statr server
		go func() {
			if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("fail to start serving: %v", err)
			}
		}()

		// todo: add order expire check for all users

		// notify signal to chan
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// wait for signal and block the app
		<-quit

		// quit signal received adn end app
		log.Println("Shutting down gateway...")

		// ctx
		cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// shutdown server
		if err := svr.Shutdown(cctx); err != nil {
			log.Fatal("Server forced to shutdown: ", err)
		}

		return nil
	},
}

// stop app
var stopCmd = &cli.Command{
	Name:  "stop",
	Usage: "stop server",
	Action: func(_ *cli.Context) error {
		pidpath, err := homedir.Expand("./")
		if err != nil {
			return nil
		}
		pd, _ := os.ReadFile(path.Join(pidpath, "pid"))
		err = kill(string(pd))
		if err != nil {
			return err
		}

		quit <- syscall.SIGTERM

		log.Println("gateway gracefully exit...")

		return nil
	},
}

// kill app
func kill(pid string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("kill", "-15", pid).Run()
	case "windows":
		return exec.Command("taskkill", "/F", "/T", "/PID", pid).Run()
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
}

// check if k8s nodes are online and set status by request to platform
func CheckOnline(platform_url string, wallet string) {
	// 创建 Kubernetes 客户端
	clientset := docker.NewK8sService()

	// 设置定时器，每隔 1 分钟查询一次
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// send request to platform every min
	for range ticker.C {
		// 获取所有节点
		nodes, err := clientset.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error fetching node list: %v\n", err)
			continue
		}

		// 打印每个节点的状态
		for _, node := range nodes.Items {
			var online bool
			// check online
			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady {
					if condition.Status == corev1.ConditionTrue {
						fmt.Printf("Node Name: %s, Online\n", node.Name)
						online = true
					} else {
						fmt.Printf("Node Name: %s, Offline\n", node.Name)
						online = false
					}
					break
				}
			}

			// get node id from label
			nid, ok := node.Labels["id"]
			if ok {
				// 尝试将标签值转换为数字
				num, err := strconv.Atoi(nid)
				if err != nil {
					logger.Info("Label value is not a valid number: %s\n", nid)
					continue
				}
				logger.Debug("node id:", num)

				// 请求平台设置节点online状态
				url := fmt.Sprintf("%s/v1/node/%s/%d/online/%v", platform_url, wallet, num, online)
				logger.Debug("url:", url)
				utils.SendPost(url)
			} else {
				logger.Debugf("No node label id=? exist for node: %s", node.Name)
			}
		}
	}
}

// check orders of this cp, and if order is end, request platform to set order status=4
func CheckOrders(platform_url string, wallet string) {
	// 设置定时器，每隔 1 分钟查询一次
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// send request to platform every min
	for range ticker.C {
		// 请求平台设置节点online状态
		url := fmt.Sprintf("%s/v1/check/order/%s", platform_url, wallet)
		fmt.Println("url:", url)
		utils.SendPost(url)
	}
}
