package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// 获取 request 中的真实客户端 ip
func requestGetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return strings.Split(r.RemoteAddr, ":")[0]
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For 可能是以","分割的地址列表
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts[0]
	}
	return hdrRealIP
}

// 当访问 localhost/healthz 时，应返回200
func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "StatusCode:", http.StatusOK)
}

// 接收客户端 request，并将 request 中带的 header 写入 response header
func returnHeader(w http.ResponseWriter, r *http.Request) {
	header := r.Header
	for key, values := range header {
		// 作业要求
		w.Header().Set(key, strings.Join(values, ""))
		// 打印到网页返回中，可以注释掉
		fmt.Fprintln(w, key, strings.Join(values, ""))
	}
	//fmt.Fprintln(w, "please check header, check if it's header contained the web content")
}

// 读取当前系统的环境变量中的 VERSION 配置，并写入 response header
func returnEnv(w http.ResponseWriter, r *http.Request) {
	version := os.Getenv("VERSION")
	// 作业要求
	fmt.Fprintln(w, "Evn parameter VERSION = ", version)
	// 打印到网页返回中，可以注释掉
	w.Header().Set("Evn parameter VERSION = ", version)
}

// Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
func printLog(w http.ResponseWriter, r *http.Request) {
	addr := requestGetRemoteAddress(r)
	// 这里 statusCode 是静态赋值的，理想状态应该是按每个客户实际请求返回的 statusCode 来返回，待完善 TODO
	statusCode := 200
	io.WriteString(w, fmt.Sprintf("Client IP: %s, ", addr))
	io.WriteString(w, fmt.Sprintf("Status Code: %d\n", statusCode))

}

// 定义函数，输入上下边界，随机输出中间任意 int 值
func randInt(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func main() {
	log.Info("http server start.")
	
	// 使用 Gauge 度量类型定义 processTime, 记录时延的瞬时快照
        processTime := prometheus.NewGauge(prometheus.GaugeOpts{
       		 Namespace: "default",
		 Name:      "http_request_processtime",
		 Help:      "The process latency time of httpserver, expect between 0 to 2s",
	     })

        // 当客户端访问 "/"时，添加 0-2 秒的随机时延, 并且将该值更新到 processTime
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	 	 delay := randInt(0, 2000)
	 	 time.Sleep(time.Millisecond * time.Duration(delay))
	 	 io.WriteString(w, fmt.Sprintf("Delay Time: %d ms\n", delay))
	 	 processTime.Set(float64(delay))
      	     })
	
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/returnHeader", returnHeader)
	http.HandleFunc("/returnEnv", returnEnv)
	http.HandleFunc("/printLog", printLog)

        // 将 processTime 注册到 metrics
	prometheus.Register(processTime)
	// 将 prometheus 的默认 handler 注册到 "/metrics"路径上
	http.Handle("/metrics", promhttp.Handler())

	go func() {
          if err := http.ListenAndServe(":80", nil); err != nil {
              log.Fatal(err)
          }
 
        }()
 
        // 优雅退出
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
        s := <-c
        log.Infof("Receive Signal [%s],Exit Properly\n", s)
}
