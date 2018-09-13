package eagletunnel

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var files = sync.Map{}

var rootPath = "./eagletunnel/http/"

func handleReq(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	switch r.Method {
	case "POST":
		handlePost(&w, r)
	case "GET":
		handleGet(&w, r)
	default:
	}
}

func handlePost(w *http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	values := r.Form
	switch path {
	case "/client.html":
		handleClientPost(values)
	case "/server.html":
	default:
	}
	config := SPrintConfig()
	writeConfig(config)
	reply, _ := readHttp("saved.html")
	(*w).Write(reply)
}

func handleClientPost(values url.Values) {
	remoteIpes := values["relayer"]
	SetRelayer(remoteIpes[0])
	localIpes := values["listen"]
	SetListen(localIpes[0])
	user_check := values["user-check"][0]
	localUser := user_check == "开启"
	if localUser {
		id := values["id"][0]
		password := values["password"][0]
		userStr := id + ":" + password
		LocalUser, _ = ParseEagleUser(userStr, nil)
	} else {
		LocalUser = nil
	}
	proxy_status := values["proxy-status"][0]
	if proxy_status == "智能" {
		PROXY_STATUS = PROXY_SMART
	} else {
		PROXY_STATUS = PROXY_ENABLE
	}
	if PROXY_STATUS == PROXY_SMART {
		whitelist_domains := values["whitelist_domains"][0]
		WhitelistDomains = strings.Split(whitelist_domains, "\r\n")
	}
	// default
	EnableHTTP = true
	EnableSOCKS5 = true
}

func handleServerPost(values url.Values) {
	localIpes := values["listen"]
	SetListen(localIpes[0])
	// default
	EnableET = true
}

func handleGet(w *http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path += "index.html"
	}
	if strings.Contains(path, ".") {
		reqType := path[strings.LastIndex(path, "."):]
		switch reqType {
		case ".css":
			(*w).Header().Set("content-type", "text/css")
		case ".js":
			(*w).Header().Set("content-type", "text/javascript")
		default:
		}
	}
	var reply []byte
	// _reply, ok := files.Load(path)
	// if ok {
	// 	reply = _reply.([]byte)
	// } else {
	// var err error
	reply, _ = readHttp(path)
	// 	if err == nil && len(reply) > 0 {
	// 		files.Store(path, reply)
	// 	}
	// }
	(*w).Write(reply)
}

func readHttp(path string) (reply []byte, err error) {
	bytes, err := ioutil.ReadFile(rootPath + path)
	if err != nil {
		return nil, err
	}
	return bytes, err
}

func writeConfig(config string) {
	ioutil.WriteFile("eagle-tunnel.conf", []byte(config), 0644)
}

func StartUI() error {
	fmt.Println("start ui at: 0.0.0.0:9090")
	http.HandleFunc("/", handleReq)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
	return err
}
