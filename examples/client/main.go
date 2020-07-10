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
	resp, err := http.Get(fmt.Sprintf("http://%s/add/tcptunnel?localPort=%d", serverAddr, port))
	if err != nil {
		return "", err
	}
	var v struct {
		Port       int    `json:"port"`
		Host       string `json:"host"`
		Identifier string `json:"identifier"`
		LocalPort  int    `json:"localPort"`
	}
	// log.Println(resp.)
	content, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

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
