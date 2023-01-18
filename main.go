package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

var name = flag.String("name", "", "请输入书名，宁可少字不能错字")
var proxyList = flag.String("proxy", "", "请输入代理文件路径，强烈建议使用代理")

func main() {
	SetupCloseHandler()
	flag.Parse()
	if *name != "" {
		var proxys []string
		if *proxyList != "" {
			proxys = readProxy(*proxyList)
		}
		var bookList []BookModel
		for i := 0; i < 10; i++ {
			searchUrl := fmt.Sprintf("https://souxs.leeyegy.com/search.aspx?key=%s&siteid=app2&page=%d", url.QueryEscape(*name), i)
			body := sendRequest(searchUrl, proxys)
			if len(body) > 0 {
				var p ListModel
				_ = json.Unmarshal([]byte(body), &p)
				if len(p.Data) == 0 {
					break
				} else {
					bookList = append(bookList, p.Data...)
				}
			} else {
				break
			}
		}
		for i := 0; i < len(bookList); i++ {
			fmt.Println(fmt.Sprintf("%d %s %s", i, bookList[i].Name, bookList[i].Author))
		}
		num := checkInput(len(bookList))
		info := bookList[num]
		var sort int
		if s, err := strconv.ParseFloat(info.Id, 64); err == nil {
			sort = int(math.Ceil(s / 1000.0))
		}
		baseUrl := fmt.Sprintf("https://downbakxs.apptuxing.com/BookFiles/Html/%d/%s/", sort, info.Id)
		body := sendRequest(baseUrl, proxys)
		if len(body) > 0 {
			temp := strings.Map(func(r rune) rune {
				if unicode.IsPrint(r) {
					return r
				}
				return -1
			}, string(body))
			temp = strings.Replace(temp, ",]", "]", -1)
			temp = strings.Replace(temp, ",}", "}", -1)
			var p BookEnum
			err := json.Unmarshal([]byte(temp), &p)
			if len(p.Data.Name) == 0 {
				fmt.Println("JSON解析失败", err)
				return
			}
			filePath := fmt.Sprintf("./%s.txt", p.Data.Name)
			file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println("文件打开失败", err)
				return
			}
			//及时关闭file句柄
			defer file.Close()
			//写入文件时，使用带缓存的 *Writer
			write := bufio.NewWriter(file)
			for i := 0; i < len(p.Data.List); i++ {
				chapter := p.Data.List[i]
				for j := 0; j < len(chapter.List); j++ {
					section := chapter.List[j]
					if section.HasContent == 1 {
						var content Content
						curlResult := string(sendRequest(fmt.Sprintf("%s/%d.html", baseUrl, section.Id), proxys))
						if len(curlResult) == 0 {
							fmt.Println("下载失败，内容为空")
							return
						}
						start := strings.Index(curlResult, "{")
						end := strings.LastIndexAny(curlResult, "}") + 1
						err2 := json.Unmarshal([]byte(curlResult[start:end]), &content)
						if err2 != nil {
							fmt.Println("JSON解析失败", err2)
							break
						}
						if content.Data.Cname == "" && content.Data.Content == "" {
							fmt.Println("Content 为空")
							break
						}
						write.WriteString(fmt.Sprintf("%s\n", content.Data.Cname))
						write.WriteString(fmt.Sprintf("%s\n", content.Data.Content))
						write.Flush()
						fmt.Println(content.Data.Cname)
					}
				}
			}
		}
	} else {
		fmt.Println("书名不能为空")
	}
}

func checkInput(count int) int {
	var num int
	fmt.Println("请输入要下载的序号：")
	fmt.Scanln(&num)
	if count < num || num < 0 {
		return checkInput(count)
	}
	return num
}

func sendRequest(path string, proxys []string) []byte {
	var client http.Client
	var proxy string
	if proxys != nil {
		proxy = proxys[rand.Intn(len(proxys)-1)]
		uri, _ := url.Parse(proxy)
		client = http.Client{
			Transport: &http.Transport{
				Proxy:               http.ProxyURL(uri),
				TLSHandshakeTimeout: 2 * time.Second,
			},
			Timeout: 5 * time.Second,
		}
	} else {
		client = http.Client{
			Timeout: 5 * time.Second,
		}
	}
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "zh-Hans-CN;q=1")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("User-Agent", "")
	resp, err := client.Do(req)
	if err != nil {
		if proxys != nil {
			var newProxy []string
			for i := 0; i < len(proxys); i++ {
				if proxys[i] != proxy {
					newProxy = append(newProxy, proxys[i])
				}
			}
			return sendRequest(path, newProxy)
		}
		time.Sleep(5 * time.Second)
		return sendRequest(path, proxys)
	}
	if resp != nil {
		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			reader.Close()
		default:
			reader = resp.Body
		}
		body, _ := ioutil.ReadAll(reader)
		defer resp.Body.Close()
		return body
	} else {
		fmt.Println(err)
		return nil
	}
}

func SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()
}

//func curl(path string, proxys []string) string {
//	var cmd *exec.Cmd
//	proxy := proxys[rand.Intn(len(proxys)-1)]
//	if proxys != nil {
//		cmd = exec.Command("curl", "-k", "-m", "5", "--proxy", proxy, path)
//	} else {
//		cmd = exec.Command("curl", "-k", "-m", "5", path)
//	}
//	out, err := cmd.Output()
//	if err != nil {
//		if proxys != nil {
//			var newProxy []string
//			for i := 0; i < len(proxys); i++ {
//				if proxys[i] != proxy {
//					newProxy = append(newProxy, proxys[i])
//				}
//			}
//			return curl(path, newProxy)
//		}
//		return ""
//	}
//	return string(out)
//}

func readProxy(path string) []string {
	var result []string
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
		return result
	}
	//及时关闭file句柄
	defer file.Close()
	//读原来文件的内容，并且显示在终端
	reader := bufio.NewReader(file)
	for {
		str, err1 := reader.ReadString('\n')
		if err1 == io.EOF {
			break
		}
		result = append(result, strings.Replace(str, "\n", "", -1))
	}
	return result
}

type ListModel struct {
	Status int         `json:"status"`
	Info   string      `json:"info"`
	Data   []BookModel `json:"data"`
}

type BookModel struct {
	Id            string `json:"Id"`
	Name          string `json:"Name"`
	Author        string `json:"Author"`
	Img           string `json:"Img"`
	Desc          string `json:"Desc"`
	BookStatus    string `json:"BookStatus"`
	LastChapterId string `json:"LastChapterId"`
	LastChapter   string `json:"LastChapter"`
	CName         string `json:"CName"`
	UpdateTime    string `json:"UpdateTime"`
	WeekHitCount  string `json:"weekHitCount"`
	MonthHitCount string `json:"monthHitCount"`
	HitCount      string `json:"hitCount"`
}

type BookEnum struct {
	Status int    `json:"status"`
	Info   string `json:"info"`
	Data   Enum   `json:"data"`
}

type Enum struct {
	Id   int       `json:"id"`
	Name string    `json:"name"`
	List []Chapter `json:"list"`
}

type Chapter struct {
	Name string    `json:"name"`
	List []Section `json:"list"`
}

type Section struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	HasContent int    `json:"hasContent"`
}

type Content struct {
	Status int    `json:"status"`
	Info   string `json:"info"`
	Data   struct {
		Id         int    `json:"id"`
		Name       string `json:"name"`
		Cid        int    `json:"cid"`
		Cname      string `json:"cname"`
		Pid        int    `json:"pid"`
		Nid        int    `json:"nid"`
		Content    string `json:"content"`
		HasContent int    `json:"hasContent"`
	} `json:"data"`
}
