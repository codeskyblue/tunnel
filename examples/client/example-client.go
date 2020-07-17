package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/koding/tunnel"
)

func requestIdentifier(serverURL string, port int) (identifier string, err error) {
	defer http.DefaultClient.CloseIdleConnections()

	url := fmt.Sprintf("%s/ktunnel/add?localPort=%d", serverURL, port)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var v struct {
		Port       int    `json:"port"`
		Host       string `json:"host"`
		Identifier string `json:"identifier"`
		LocalPort  int    `json:"localPort"`
	}
	content, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("StatusCode expect 200, got %d", resp.StatusCode)
	}

	json.Unmarshal(content, &v)

	if v.LocalPort != port {
		return "", fmt.Errorf("expect localPort %d but got %d", port, v.LocalPort)
	}

	fmt.Printf("TunnelAddress: %s:%d\n", v.Host, v.Port)
	// fmt.Printf("TestWithNetcat: nc %s %d\n", v.Host, v.Port)

	return v.Identifier, nil
}

func requestRemoveIdentifier(serverURL string, identifier string) error {
	req, _ := http.NewRequest("DELETE", serverURL+"/ktunnel/del?identifier="+identifier, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Request remove identifier failed: %v", resp.Status)
	}
	return nil
}

func main() {
	var port int
	var serverAddr string
	var debug bool
	var identifier string
	var fSecure bool

	flag.StringVar(&serverAddr, "server", "localhost:5000", "server address")
	flag.IntVar(&port, "port", 3000, "published port")
	flag.BoolVar(&debug, "debug", false, "show debug log")
	flag.StringVar(&identifier, "ident", "", "identifier (optional)")
	flag.BoolVar(&fSecure, "secure", false, "enable https mode")
	flag.Parse()

	serverURL := "http://" + serverAddr
	if fSecure {
		serverURL = "https://" + serverAddr
	}

	if identifier == "" {
		if port == 0 {
			log.Fatal("-port must be specified")
		}
		fmt.Printf("TunnelServer: %s\n", serverAddr)
		fmt.Printf("LocalPort: %d\n", port)

		var err error
		identifier, err = requestIdentifier(serverURL, port)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Identifier:", identifier)

	dialFunc := func(network, address string) (net.Conn, error) {
		log.Println("Dial", network, address)
		address += ":443"
		conn, err := tls.Dial(network, address, nil)
		if err != nil {
			log.Println("Tls dial error:", err)
		}
		return conn, err
	}

	config := &tunnel.ClientConfig{
		Identifier: identifier,
		ServerAddr: serverAddr,
		Debug:      debug,
	}

	if fSecure {
		config.Dial = dialFunc
	}
	client, err := tunnel.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(c chan os.Signal, identifier string) {
		// Block until a signal is received.
		s := <-c
		fmt.Println("Got signal:", s)
		requestRemoveIdentifier(serverURL, identifier)
		os.Exit(0)
	}(c, identifier)
	client.Start()
}
