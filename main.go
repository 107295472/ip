package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"github.com/tidwall/gjson"
	"yin/ip/crontask"
	"yin/ip/util"
)

func main() {
	crontask.CronTask()
	runtime.GOMAXPROCS(runtime.NumCPU())
	datFile := flag.String("qqwry",util.GetCurrentPath()+"/qqwry.dat", "纯真 IP 库的地址")
	port := flag.String("port",GetKey("port"), "HTTP 请求监听端口号")
	flag.Parse()

	IPData.FilePath = *datFile
	startTime := time.Now().UnixNano()
	res := IPData.InitIPData()

	if v, ok := res.(error); ok {
		log.Panic(v)
	}
	endTime := time.Now().UnixNano()
	log.Printf("IP 库加载完成 共加载:%d 条 IP 记录, 所花时间:%.1f ms\n", IPData.IPNum, float64(endTime-startTime)/1000000)

	// 下面开始加载 http 相关的服务
	http.HandleFunc("/", findIP)

	log.Printf("开始监听网络端口:%s", *port)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil); err != nil {
		log.Println(err)
	}
}

// findIP 查找 IP 地址的接口
func findIP(w http.ResponseWriter, r *http.Request) {
	res := NewResponse(w, r)
	ipStr:=r.Form.Get("ip")
	ti:=r.Form.Get("time")
	if len(ipStr)==0 {
		ipStr = r.RemoteAddr
		index:=strings.LastIndexByte(ipStr, ':')
		ipStr=util.Substr(ipStr,0,index)
		//fmt.Println(ip)
	}
	token := r.Form.Get("token")
	h := md5.New()
	h.Write([]byte(ti+GetKey("key")))
	cipherStr := h.Sum(nil)
	tok:=hex.EncodeToString(cipherStr)
	if tok!=strings.ToLower(token) {
		res.ReturnError(http.StatusBadRequest, 200002, "签名错误")
		return
	}
	//if ipStr == "" {
	//	res.ReturnError(http.StatusBadRequest, 200001, "请填写 IP 地址")
	//	return
	//}

	//ips := strings.Split(ip, ",")

	qqWry := NewQQwry()

	rs := ResultQQwry{}
	if len(ipStr) > 0 {
		rs = qqWry.Find(ipStr)
		//for _, v := range ips {
		//
		//}
	}
	res.ReturnSuccess(rs)
}
func GetKey(key string) string{
	dat, err := ioutil.ReadFile(util.GetCurrentPath()+"/config.json")
	check(err)
	value := gjson.Get(string(dat), key)
	return value.Str
}
//异常处理
func check(e error) {
	if e != nil {
		panic(e)
	}
}
