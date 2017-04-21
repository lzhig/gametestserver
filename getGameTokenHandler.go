package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/golang/glog"
)

type getGameTokenHandlerT struct {
	webClient *http.Client
	config    *configT
}

type requestT struct {
	Username string `json:"username"`
	Password string `json:"password"`
	AppID    string `json:"appid"`
}

type userinfoT struct {
	UID         uint32 `json:"uid"`
	Username    string `json:"username"`
	AppID       string `json:"app_id"`
	AppName     string `json:"app_name"`
	Timestamp   int64  `json:"timestamp"`
	UserIP      string `json:"user_ip"`
	AccessToken string `json:"access_token"`
	Extra       string `json:"extra"`
	GuideStatus int    `json:"guide_status"`
}

type responseT struct {
	Ret      uint32       `json:"ret"`
	UserInfo userinfoT    `json:"userinfo"`
	Balance  balanceInfoT `json:"balance"`
}

func (handler getGameTokenHandlerT) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	/*
		err := r.ParseForm()
		if err != nil {
			io.WriteString(w, fmt.Sprint("error:", err))
			return
		}
	*/
	if strings.ToUpper(r.Method) != "POST" {
		io.WriteString(w, fmt.Sprintf("error: please use POST. method: %s", r.Method))
		return
	}

	postData, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	//fmt.Println(string(postData))

	var requestData requestT
	if err := json.Unmarshal(postData, &requestData); err != nil {
		io.WriteString(w, fmt.Sprint("error:", err))
		return
	}

	if err := handler.initialize(); err != nil {
		io.WriteString(w, fmt.Sprint("error:", err))
		return
	}

	//fmt.Println(requestData)
	if err := handler.login(requestData.Username, requestData.Password); err != nil {
		io.WriteString(w, fmt.Sprint("error:", err))
		return
	}

	var responseData responseT
	_appid, _appname, _userip, _uid, _username, _timestamp, _extra, _accessToken, err := handler.GetGameInfo(requestData.AppID)
	if err != nil {
		io.WriteString(w, fmt.Sprint("error:", err))
		return
	}

	responseData.UserInfo.UID = _uid
	responseData.UserInfo.Username = _username
	responseData.UserInfo.Timestamp = _timestamp
	responseData.UserInfo.Extra = _extra
	responseData.UserInfo.AccessToken = _accessToken
	responseData.UserInfo.AppID = _appid
	responseData.UserInfo.AppName = _appname
	responseData.UserInfo.UserIP = _userip

	_cash, _nm, _coin, err := handler.GetBalance()
	if err != nil {
		io.WriteString(w, fmt.Sprint("error:", err))
		return
	}

	responseData.Balance.Cash = _cash
	responseData.Balance.Coin = _coin
	responseData.Balance.Nm = _nm

	responseData.Ret = 0

	resp, err := json.Marshal(responseData)
	if err != nil {
		io.WriteString(w, fmt.Sprint("error:", err))
		return
	}

	w.Write(resp)
}

func (handler *getGameTokenHandlerT) initialize() error {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return err
	}
	handler.webClient = &http.Client{Jar: jar}

	return nil
}

type hashInfoT struct {
	Key string `json:"key"`
	IV  string `json:"iv"`
}

func (handler *getGameTokenHandlerT) getHashInfo() (*hashInfoT, error) {
	hashInfoResp, err := handler.webClient.Get(fmt.Sprintf("http://%s/%s", handler.config.PlatformURL, handler.config.InterfaceHashInfo))
	if err != nil {
		return nil, err
	}
	hashInfoData, err := ioutil.ReadAll(hashInfoResp.Body)
	hashInfoResp.Body.Close()
	if err != nil {
		return nil, err
	}
	//glog.Info("data:", string(hashInfoData))

	hashInfo := &hashInfoT{}
	err = json.Unmarshal(hashInfoData, &hashInfo)
	if err != nil {
		return nil, err
	}

	return hashInfo, nil
}

func (handler *getGameTokenHandlerT) login(username string, password string) error {
	hashInfo, err := handler.getHashInfo()
	if err != nil {
		return err
	}

	block, err := aes.NewCipher([]byte(hashInfo.Key))
	if err != nil {
		return err
	}

	mode := cipher.NewCBCEncrypter(block, []byte(hashInfo.IV))
	dst := make([]byte, aes.BlockSize)
	passwordTmp := make([]byte, 16)
	copy(passwordTmp, password)
	//glog.Info("password:", passwordTmp)
	mode.CryptBlocks(dst, passwordTmp)
	//glog.Info("dst:", dst)

	encoded := base64.StdEncoding.EncodeToString(dst)
	//glog.Info(encoded)

	postdata := url.Values{
		"username":          []string{username},
		"password":          []string{encoded},
		"remember_password": []string{"0"},
		"verify_code":       []string{""},
	}

	resp, err := handler.webClient.PostForm(fmt.Sprintf("http://%s/%s", handler.config.PlatformURL, handler.config.InterfaceLogin), postdata)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return err
	}
	//glog.Info("data:", string(data))

	loginResult := make(map[string]string)
	err = json.Unmarshal(data, &loginResult)
	if err != nil {
		return err
	}
	//glog.Info(loginResult)
	if _, ok := loginResult["error_code"]; ok {
		// 注册用户
		postdata := url.Values{
			"reg_username":        []string{username},
			"reg_password":        []string{encoded},
			"reg_password_repeat": []string{encoded},
			"new_reg_source":      []string{"0"},
		}
		resp, err := handler.webClient.PostForm(fmt.Sprintf("http://%s/%s", handler.config.PlatformURL, handler.config.InterfaceRegister), postdata)
		if err != nil {
			return err
		}

		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return err
		}
		glog.Info("data:", string(data))
	}

	return nil
}

type gameInfoT struct {
	AppID       string `json:"app_id"`
	AppName     string `json:"app_name"`
	UID         uint32 `json:"uid"`
	Username    string `json:"username"`
	Timestamp   int64  `json:"timestamp"`
	UserIP      string `json:"user_ip"`
	AccessToken string `json:"access_token"`
	Extra       string `json:"extra"`
}

func (handler *getGameTokenHandlerT) GetGameInfo(gameAppID string) (string, string, string, uint32, string, int64, string, string, error) {

	gameInfoResp, err := handler.webClient.Get(fmt.Sprintf("http://%s/%s/%s", handler.config.PlatformURL, handler.config.InterfaceGameInfo, gameAppID))
	if err != nil {
		return "", "", "", 0, "", 0, "", "", err
	}
	gameInfoData, err := ioutil.ReadAll(gameInfoResp.Body)
	gameInfoResp.Body.Close()
	if err != nil {
		return "", "", "", 0, "", 0, "", "", err
	}
	//glog.Info("gameinfo data:", string(gameInfoData))

	var t interface{}
	err = json.Unmarshal(gameInfoData, &t)
	if err != nil {
		return "", "", "", 0, "", 0, "", "", err
	}
	m, ok := t.(map[string]interface{})
	if ok {
		v, r := m["error"]
		if r {
			return "", "", "", 0, "", 0, "", "", errors.New(fmt.Sprint(v))
		}
	}

	var gameInfo gameInfoT
	err = json.Unmarshal(gameInfoData, &gameInfo)
	if err != nil {
		return "", "", "", 0, "", 0, "", "", err
	}

	return gameInfo.AppID, gameInfo.AppName, gameInfo.UserIP, gameInfo.UID, gameInfo.Username, gameInfo.Timestamp, gameInfo.Extra, gameInfo.AccessToken, nil
}

type balanceInfoT struct {
	Cash uint64 `json:"cash"`
	Coin uint64 `json:"coin"`
	Nm   uint64 `json:"nm"`
}
type jsonBalanceT struct {
	Code int          `json:"code"`
	Info balanceInfoT `json:"info"`
}

func (handler *getGameTokenHandlerT) GetBalance() (cash uint64, nm uint64, coin uint64, err error) {
	balanceResp, err := handler.webClient.Get(fmt.Sprintf("http://%s/%s", handler.config.PlatformURL, handler.config.InterfaceBalance))
	if err != nil {
		return 0, 0, 0, err
	}
	balanceData, err := ioutil.ReadAll(balanceResp.Body)
	balanceResp.Body.Close()
	if err != nil {
		return 0, 0, 0, err
	}
	glog.Info("balance data:", string(balanceData))
	var balance jsonBalanceT
	err = json.Unmarshal(balanceData, &balance)
	if err != nil || balance.Code != 0 {
		glog.Error(balance)
		return 0, 0, 0, err
	}

	return balance.Info.Cash, balance.Info.Nm, balance.Info.Coin, nil
}
