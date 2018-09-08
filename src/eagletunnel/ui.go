package eagletunnel

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

var files = sync.Map{}

var rootPath = "./eagletunnel/http/"

func hello(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	path := r.URL.Path
	if path == "/" {
		path += "index.html"
	}
	if strings.Contains(path, ".") {
		reqType := path[strings.LastIndex(path, "."):]
		switch reqType {
		case ".css":
			w.Header().Set("content-type", "text/css")
		case ".js":
			w.Header().Set("content-type", "text/javascript")
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
	w.Write(reply)
}

func readHttp(path string) (reply []byte, err error) {
	bytes, err := ioutil.ReadFile(rootPath + path)
	if err != nil {
		return nil, err
	}
	if path == "/main.css" {
		path = ""
	}
	return bytes, err
}

func StartUI() error {
	http.HandleFunc("/", hello)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
	return err
}
