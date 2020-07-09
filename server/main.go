package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/koding/tunnel"
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

func main() {
	rand.Seed(time.Now().UnixNano())

	log.Println("Tunnel Server starts")

	server, err := tunnel.NewServer(&tunnel.ServerConfig{
		Debug: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/add/tcptunnel", func(w http.ResponseWriter, r *http.Request) {
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
			"host":       "localhost",
			"localPort":  localPort,
			"identifier": identifier,
		})
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
		w.Write(jsonData)
		// server.DeleteAddr()
	})
	http.Handle("/", server)

	http.ListenAndServe(":5000", nil)
}
