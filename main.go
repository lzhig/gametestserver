package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/golang/glog"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	// 读取配置
	config, err := loadConfig("config.json")
	if err != nil {
		glog.Error("Failed to load config.json. Error: ", err)
		return
	}

	// 读取crossdomain.xml
	crossdomain, err := ioutil.ReadFile("crossdomain.xml")
	if err != nil {
		glog.Error("Failed to load crossdomain.xml. Error: ", err)
		return
	}

	http.Handle("/GetGameToken", getGameTokenHandlerT{config: config})

	http.HandleFunc("/crossdomain.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Write(crossdomain)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.HTTPServicePort), nil); err != nil {
		glog.Error("Failed to start webservice. Error: ", err)
		return
	}
}

type configT struct {
	HTTPServicePort   uint16 `json:"http_service_port"`
	PlatformURL       string `json:"platform_url"`
	InterfaceHashInfo string `json:"interface_hashinfo"`
	InterfaceLogin    string `json:"interface_login"`
	InterfaceRegister string `json:"interface_register"`
	InterfaceGameInfo string `json:"interface_gameinfo"`
	InterfaceBalance  string `json:"interface_balance"`
}

func loadConfig(file string) (*configT, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cfg := &configT{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
