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

func main() {
	var port int
	var serverAddr string
	var identifier string

	flag.StringVar(&serverAddr, "server", "localhost:5000", "server address")
	flag.IntVar(&port, "port", 0, "local port")
	flag.StringVar(&identifier, "ident", "", "identifier")
	flag.Parse()

	log.Printf("TS: %s, local-port: %d", serverAddr, port)

	resp, err := http.Get(fmt.Sprintf("http://%s/add/tcptunnel", serverAddr))
	if err != nil {
		panic(err)
	}
	var v struct {
		Port       int    `json:"port"`
		Host       string `json:"host"`
		Identifier string `json:"identifier"`
	}
	// log.Println(resp.)
	content, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	json.Unmarshal(content, &v)
	log.Println("Close request")

	log.Printf("Identifer: %s", v.Identifier)
	log.Printf("Remote address: %s:%d", v.Host, v.Port)

	http.DefaultClient.CloseIdleConnections()
	// logging.NewLogger("tcp").SetLevel(logging.DEBUG)

	client, err := tunnel.NewClient(&tunnel.ClientConfig{
		Identifier: v.Identifier,
		// FetchIdentifier: func() (string, error)
		ServerAddr: serverAddr,
		Debug:      true,
	})
	if err != nil {
		log.Fatal(err)
	}
	client.Start()
}
