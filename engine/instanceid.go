package engine

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

type InstanceID struct {
	Id string `json:"id"`
	Cloud string `json:"cloud"`
	Addr string `json:"addr"`
}

type revealFunction func () (*InstanceID, error)

func dummy() (*InstanceID, error) {
	return &InstanceID{Id : "", Cloud : ""}, nil
}

// Get preferred outbound ip of this machine
func getOutboundIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()
}

func digitalOcean() (*InstanceID, error) {
	var resp *http.Response
	var err error
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err = client.Get("http://169.254.169.254/metadata/v1/id")

	if err != nil || resp.StatusCode != 200 {
		return nil, err
	}

	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &InstanceID{Id : string(body), Cloud : "DigitalOcean"}, nil
}

func aws() (*InstanceID, error) {
	var resp *http.Response
	var err error
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err = client.Get("http://169.254.169.254/latest/meta-data/instance-id")

	if err != nil || resp.StatusCode != 200{
		return nil, err
	}

	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &InstanceID{Id : string(body), Cloud : "AWS"}, nil
}

func GetInstanceID() *InstanceID {
	functors := []revealFunction {
		digitalOcean,
		aws,
		dummy,
	}

	for _, functor := range functors {
		result, err := functor()
		if err == nil && result != nil {
			result.Addr = getOutboundIP()
			return result
		}
	}
	return nil
}

