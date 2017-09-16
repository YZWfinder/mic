package mic

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	GET  = "GET"
	POST = "POST"
	GIF  = "image/gif"
	PNG  = "image/png"
	JPG  = "image/jpeg"
	ICO  = "image/x-icon"
	PDF  = "application/pdf"
	CSS  = "text/css;charset=utf-8"
	HTML = "text/html;charset=utf-8"
	TXT  = "text/plain;charset=utf-8"
	JS   = "text/javascript;charset=utf-8"
)

//常用的mime类型
const (
	applicationJson = "application/json"
	applicationXml  = "application/xml"
	textXml         = "text/xml"
)

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
}

type RouteMux struct {
	routes  []*route
	filters []http.HandlerFunc
}

func Server() *RouteMux {
	return &RouteMux{}
}
func (m *RouteMux) Run(port string) {
	m.Filter(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "POST, GET")
		h.Add("Access-Control-Allow-Headers", "Content-Type")
		h.Set("Content-Type", "application/json;charset=utf-8")
	})
	http.ListenAndServe(port, m)

}
func (m *RouteMux) Get(pattern string, handler http.HandlerFunc) {
	m.AddRoute(GET, pattern, handler)
}
func (m *RouteMux) Post(pattern string, handler http.HandlerFunc) {
	m.AddRoute(POST, pattern, handler)
}

func (m *RouteMux) Public(pattern string) {
	dir, _ := os.Getwd()
	//将正则表达式附加到param，以匹配所有内容 //这是在前缀之后
	pattern = pattern + "(.+)"
	m.AddRoute(GET, pattern, func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Path)
		path = filepath.Join(dir, path)
		ur := strings.Split(r.URL.Path, "/")
		//ur := strings.Split(r.RequestURI, "/")
		url := ur[len(ur)-1]
		urls := strings.Split(url, ".")[1]
		switch urls {
		case "html":
			w.Header().Set("Content-Type", HTML)
		case "png":
			w.Header().Set("Content-Type", PNG)
		case "css":
			w.Header().Set("Content-Type", CSS)
		case "jpg":
			w.Header().Set("Content-Type", JPG)
		case "js":
			w.Header().Set("Content-Type", JS)
		case "gif":
			w.Header().Set("Content-Type", GIF)
		case "ico":
			w.Header().Set("Content-Type", ICO)
		case "pdf":
			w.Header().Set("Content-Type", PDF)
		case "txt":
			w.Header().Set("Content-Type", TXT)
		default:
			w.Header().Set("Content-Type", "text/xml;charset=utf-8")
		}
		http.ServeFile(w, r, path)
	})
}

func (m *RouteMux) AddRoute(method string, pattern string, handler http.HandlerFunc) {
	regex, regexErr := regexp.Compile(pattern)
	if regexErr != nil {
		//TODO add error handling here to avoid panic
		panic(regexErr)
		return
	}
	//现在创建的路线
	route := &route{}
	route.method = method
	route.regex = regex
	route.handler = handler

	//最后添加到路由列表
	m.routes = append(m.routes, route)
}

// 过滤器添加了中间件过滤器。
func (m *RouteMux) Filter(filter http.HandlerFunc) {
	m.filters = append(m.filters, filter)
}

//需要的http。处理程序接口。这个方法是由
//http服务器将处理所有页面路由
func (m *RouteMux) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path

	//包装响应写入器，在我们的自定义接口中
	w := &responseWriter{writer: rw}
	//find a matching Route
	for _, route := range m.routes {

		//如果方法不匹配，跳过这个处理程序
		//我。如果请求。方法是“把”路线。方法必须“把”
		if r.Method != route.method {
			continue
		}

		//检查路由模式是否匹配url
		if !route.regex.MatchString(requestPath) {
			continue
		}

		//执行中间件过滤器
		for _, filter := range m.filters {
			filter(w, r)
			if w.started {
				return
			}
		}

		//调用请求处理程序
		route.handler(w, r)
		break
	}

	//如果不匹配url，抛出一个未发现的异常
	if w.started == false {
		http.NotFound(w, r)
	}
}

// -----------------------------------------------------------------------------
//简单包装器围绕一个ResponseWriter
//responseWriter是http.responseWriter的一个包装器
//跟踪响应是否被写入。它还允许我们
//自动设置某些标头，例如content-type，
// Access-Control-Allow-Origin等等。
type responseWriter struct {
	writer  http.ResponseWriter
	started bool
	status  int
}

//Header返回由WriteHeader发送的标题映射。
func (w *responseWriter) Header() http.Header {
	return w.writer.Header()
}

//Write将数据写入到连接，作为HTTP应答的一部分，
//集合开始为真
func (w *responseWriter) Write(p []byte) (int, error) {
	w.started = true
	return w.writer.Write(p)
}

//WriteHeader发送带有状态代码的HTTP响应头，
//集合开始为真
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.started = true
	w.writer.WriteHeader(code)
}

// -----------------------------------------------------------------------------
//struct转json 发送
func ServeJson(w http.ResponseWriter, v interface{}) {
	//content, err := json.Marshal(v, "", "  ")
	content, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", applicationJson)
	w.Write(content)
}

//读json转struct
func ReadJson(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

//struct转xml 发送
func ServeXml(w http.ResponseWriter, v interface{}) {
	content, err := xml.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(content)
}

//读xml转struct
func ReadXml(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, v)
}

//读string
func ReadString(r *http.Request) (string, error) {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return "", err
	}
	return string(body), nil
}
