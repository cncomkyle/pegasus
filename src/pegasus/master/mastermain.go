package main

import (
	"fmt"
	"net/http"
	"pegasus/cfgmgr"
	"pegasus/log"
	"pegasus/route"
	"pegasus/server"
	"pegasus/uri"
	"pegasus/util"
	"strings"
)

var cfgServerIP = "127.0.0.1"

type Master struct {
	Ip           string
	ListenPort   int
	masterServer *server.Server
	masterAddr   string
}

var masterSelf = new(Master)

func discoverIp() error {
	log.Info("Discover self ip address")
	ip, err := cfgmgr.DiscoverIpFromCfg(cfgServerIP)
	if err != nil {
		return err
	}
	masterSelf.Ip = ip
	log.Info("Discover self ip address as %s", ip)
	return nil
}

func prepareNetwork() error {
	log.Info("Prepare network stuff")
	s := new(server.Server)
	if err := discoverIp(); err != nil {
		return err
	}
	if err := s.Listen(masterSelf.Ip); err != nil {
		return fmt.Errorf("Fail to listen, %v", err)
	}
	masterSelf.masterServer = s
	masterSelf.masterAddr = s.GetListenAddr()
	_, port, err := util.SplitAddr(masterSelf.masterAddr)
	if err != nil {
		return err
	}
	masterSelf.ListenPort = port
	log.Info("Listen on %s:%d", masterSelf.Ip, masterSelf.ListenPort)
	return nil
}

func registerOnCfgServer() (err error) {
	log.Info("Register on cfg server")
	u := &util.HttpUrl{
		IP:   cfgServerIP,
		Port: cfgmgr.CfgServerPort,
		Uri:  uri.CfgMasterUri,
	}
	_, err = util.HttpPostStr(u, masterSelf.masterAddr)
	if err != nil {
		return fmt.Errorf("Fail to register, %v", err)
	}
	log.Info("Register on cfg server done")
	return
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Handle test request")
	s, err := util.HttpReadRequestTextBody(r)
	if err != nil {
		err = fmt.Errorf("Fail to read request test body, %v", err)
		log.Error(err.Error())
		server.FmtResp(w, err, "")
		return
	}
	log.Info("Get test msg as %s", s)
	ips, ports := wmgr.getAllWorkerAddr()
	workerResp := make([]string, 0)
	for i := 0; i < len(ips); i++ {
		u := &util.HttpUrl{
			IP:   ips[i],
			Port: ports[i],
			Uri:  uri.WorkerTestUri,
		}
		if resp, err := util.HttpPostStr(u, s); err != nil {
			log.Error(err.Error())
			server.FmtResp(w, err, s)
			return
		} else {
			log.Info("Get resp as %s", resp)
			workerResp = append(workerResp,
				fmt.Sprintf("%s%d: %s", ips[i], ports[i], resp))
		}
	}
	server.FmtResp(w, err, strings.Join(workerResp, "\n"))
	log.Info("Handle test request done")
}

func registerRoutes() {
	route.RegisterRoute(&route.Route{
		Name:    "registerWorkerHandler",
		Method:  http.MethodGet,
		Path:    uri.MasterRegisterWokerUri,
		Handler: registerWorkerHandler,
	})
	route.RegisterRoute(&route.Route{
		Name:    "verifyWorkerHandler",
		Method:  http.MethodPost,
		Path:    uri.MasterRegisterWokerUri,
		Handler: verifyWorkerHandler,
	})
	route.RegisterRoute(&route.Route{
		Name:    "verifyWorkerHandler",
		Method:  http.MethodPost,
		Path:    uri.MasterRegisterWokerUri,
		Handler: verifyWorkerHandler,
	})
	route.RegisterRoute(&route.Route{
		Name:    "workerHbHandler",
		Method:  http.MethodPost,
		Path:    uri.MasterWorkerHbUri,
		Handler: workerHbHandler,
	})
	route.RegisterRoute(&route.Route{
		Name:    "workerHbIntervalHandler",
		Method:  http.MethodGet,
		Path:    uri.MasterWorkerHbIntervalUri,
		Handler: workerHbIntervalHandler,
	})
	route.RegisterRoute(&route.Route{
		Name:    "taskReportHandler",
		Method:  http.MethodPost,
		Path:    uri.MasterWorkerTaskReportUri,
		Handler: taskReportHandler,
	})
	route.RegisterRoute(&route.Route{
		Name:    "testHandler",
		Method:  http.MethodPost,
		Path:    uri.MasterTestUri,
		Handler: testHandler,
	})
	route.RegisterRoute(&route.Route{
		Name:    "testRunHandler",
		Method:  http.MethodGet,
		Path:    uri.MasterTestUri,
		Handler: testRunHandler,
	})
}

func initLogger() error {
	consoleLogger := &log.ConsoleLogger{
		Level: log.LevelInfo,
	}
	if err := log.RegisterLogger(consoleLogger); err != nil {
		return err
	}
	log.Info("=====Start Master server=====")
	return nil
}

func main() {
	if err := initLogger(); err != nil {
		panic(fmt.Errorf("Fail to init logger, %v", err))
	}
	registerRoutes()
	cfgmgr.WaitForCfgServerUp(cfgServerIP)
	if err := prepareNetwork(); err != nil {
		panic(err)
	}
	if err := registerOnCfgServer(); err != nil {
		panic(err)
	}
	panic(masterSelf.masterServer.Serve())
}