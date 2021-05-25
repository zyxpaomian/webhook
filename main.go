package main

import (
	"math/rand"
	"webhook/http"
	"webhook/http/handle"
	"k8s.io/klog"
	go_http "net/http"
	"runtime"
	"time"
	"flag"
	"fmt"
	"crypto/tls"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UTC().UnixNano())

	// 初始化配置 && 日志
	var port int
	var certFile string
	var keyFile string
	flag.IntVar(&port, "port", 443, "Webhook server port.")
	flag.StringVar(&certFile, "tlsCertFile", "/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&keyFile, "tlsKeyFile", "/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()


	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
		return
	}
	// 启动HTTP服务
	mux := http.New()
	handle.InitHandle(mux)
	srv := &go_http.Server{
		Handler:      mux.GetRouter(),
		Addr:         fmt.Sprintf(":%d", port),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		WriteTimeout: 15 * time.Hour,
		ReadTimeout:  15 * time.Hour,
	}
	go srv.ListenAndServeTLS("", "")
	//go srv.ListenAndServe()

	// 等待
	select {}
}
