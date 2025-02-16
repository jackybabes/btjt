package main

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func readTorrent(filename string) []byte {
	data, _ := os.ReadFile(filename)
	return data
}

func bytesToInt(b []byte) int {
	s := string(b)
	i, _ := strconv.Atoi(s)
	return i
}

func getClient(proxy bool) *http.Client {
	// Proxy through burp
	// Define the Burp Suite proxy URL
	proxyURL, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	// Create a custom Transport with the proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		// Skip TLS verification (needed for HTTPS if Burp is using its cert)
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if proxy {
		// Create an HTTP client with the custom transport
		return &http.Client{
			Transport: transport,
		}
	}
	return &http.Client{}
}
