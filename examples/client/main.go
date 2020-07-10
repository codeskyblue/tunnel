package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/koding/tunnel"
)

func requestIdentifier(serverAddr string, port int) (identifier string, err error) {
	defer http.DefaultClient.CloseIdleConnections()

	url := "http://%s/ktunnel/add?localPort=%d"
	// url = "http://%s/add/tunnel?localPort=%d"

	resp, err := http.Get(fmt.Sprintf(url, serverAddr, port))
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

	fmt.Printf("Identifer: %q\n", v.Identifier)
	fmt.Printf("TunnelAddress: %s:%d\n", v.Host, v.Port)
	// fmt.Printf("TestWithNetcat: nc %s %d\n", v.Host, v.Port)

	return v.Identifier, nil
}

func main() {
	var port int
	var serverAddr string
	var debug bool
	var identifier string

	flag.StringVar(&serverAddr, "server", "localhost:5000", "server address")
	flag.IntVar(&port, "port", 0, "published port")
	flag.BoolVar(&debug, "debug", false, "show debug log")
	flag.StringVar(&identifier, "ident", "", "identifier (optional)")
	flag.Parse()

	if identifier == "" {
		if port == 0 {
			log.Fatal("-port must be specified")
		}
		fmt.Printf("TunnelServer: %s\n", serverAddr)
		fmt.Printf("LocalPort: %d\n", port)

		var err error
		identifier, err = requestIdentifier(serverAddr, port)
		if err != nil {
			log.Fatal(err)
		}
	}

	client, err := tunnel.NewClient(&tunnel.ClientConfig{
		Identifier: identifier,
		ServerAddr: serverAddr,
		Debug:      debug,
	})
	if err != nil {
		log.Fatal(err)
	}
	client.Start()
}
