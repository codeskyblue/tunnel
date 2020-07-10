package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/koding/tunnel"
	"github.com/koding/tunnel/proto"
)

func getFreePort() (port int, err error) {
	// return 7100, nil
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().String()
	_, portString, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(portString)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func GetRemoteAddr(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func main() {
	pAddr := flag.String("addr", ":5000", "listen address")
	pDebug := flag.Bool("debug", false, "debug mode")
	flag.Parse()

	localIpAddress := GetLocalIP()

	rand.Seed(time.Now().UnixNano())

	log.Printf("Tunnel Server starts, listen on %s", *pAddr)
	log.Printf("Current IP address: %s", localIpAddress)

	server, err := tunnel.NewServer(&tunnel.ServerConfig{
		Debug: *pDebug,
	})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/ktunnel/status", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "IMOK")
	})

	http.HandleFunc("/ktunnel/add", func(w http.ResponseWriter, r *http.Request) {
		freePort, err := getFreePort()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		localPort, err := strconv.Atoi(r.FormValue("localPort"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// localPort := 27015
		identifier := randSeq(10)
		lis, err := net.Listen("tcp", ":"+strconv.Itoa(freePort))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		server.AddAddr(lis, localPort, nil, identifier)
		jsonData, _ := json.Marshal(map[string]interface{}{
			"port":       freePort,
			"host":       localIpAddress,
			"localPort":  localPort,
			"identifier": identifier,
		})

		remoteAddr := GetRemoteAddr(r)
		rHost, _, _ := net.SplitHostPort(remoteAddr)

		log.Printf("New tunnel request from %s:%d", rHost, localPort)

		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
		w.Write(jsonData)
		// server.DeleteAddr()
	})
	http.HandleFunc(proto.ControlPath, server.ServeHTTP) // path:/ktunnel/_controlPath/

	http.ListenAndServe(*pAddr, nil)
}
