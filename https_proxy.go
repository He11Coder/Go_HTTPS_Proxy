package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"proxy/repository"
)

type Proxy struct {
	hostKey      []byte
	requestRepo  repository.IRequestRepository
	responseRepo repository.IResponseRepository
}

func createProxy(reqRepo repository.IRequestRepository, respRepo repository.IResponseRepository) (*Proxy, error) {
	certKey, err := os.ReadFile("cert.key")
	if err != nil {
		return nil, err
	}

	return &Proxy{hostKey: certKey, requestRepo: reqRepo, responseRepo: respRepo}, nil
}

func changeURLToTarget(r *http.Request, target string) error {
	if !strings.HasPrefix(target, "https") {
		target = "https://" + target
	}

	newURL, err := url.Parse(target)
	if err != nil {
		return err
	}

	newURL.Path = r.URL.Path
	newURL.RawQuery = r.URL.RawQuery
	r.URL = newURL
	r.RequestURI = ""

	return nil
}

func copyHeader(in, out http.Header) {
	for key, value := range in {
		for _, v := range value {
			out.Add(key, v)
		}
	}
}

func (p *Proxy) httpProxy(w http.ResponseWriter, r *http.Request) {
	if r.URL.Scheme != "http" {
		msg := "unsupported protocol: " + r.URL.Scheme
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	r.Header.Del("Proxy-Connection")
	r.Header.Set("Connection", "close")

	r.RequestURI = ""

	reqID, err := p.requestRepo.AddRequest(r)
	if err != nil {
		fmt.Println("error while dumping the request:", err)
		return
	}

	response, err := client.Do(r)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	_, err = p.responseRepo.AddResponse(response, reqID)
	if err != nil {
		fmt.Println("error while dumping the response:", err)
		return
	}

	copyHeader(response.Header, w.Header())
	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}

func (p *Proxy) httpsProxy(w http.ResponseWriter, r *http.Request) {
	/*if r.URL.Scheme != "https" {
		msg := "unsupported protocol: " + r.URL.Scheme
		http.Error(w, msg, http.StatusBadRequest)

		return
	}*/

	w.WriteHeader(http.StatusOK)
	hj, ok := w.(http.Hijacker)
	if !ok {
		fmt.Println("http server doesn't support hijacking connection")
		return
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		fmt.Println("http hijacking failed")
		return
	}

	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		fmt.Println("error while splitting host and port")
		return
	}

	cmd := exec.Command("./cert_gen.sh", host, strconv.FormatInt(rand.Int63(), 10))
	hostCert, err := cmd.Output()
	if err != nil {
		fmt.Println("error while generating host sertificate:", err, hostCert)
		return
	}

	tlsCert, err := tls.X509KeyPair(hostCert, p.hostKey)
	if err != nil {
		fmt.Println("error while creating tls certificate:", err)
		return
	}

	_, err = clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	if err != nil {
		fmt.Println("failed to send status to the client:", err)
		return
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:               tls.VersionTLS10,
		MaxVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{tlsCert},
	}

	tlsConn := tls.Server(clientConn, tlsConfig)
	defer tlsConn.Close()

	connReader := bufio.NewReader(tlsConn)

	for {
		req, err := http.ReadRequest(connReader)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("error while reading the client request:", err)
			return
		}

		err = changeURLToTarget(req, r.Host)
		if err != nil {
			fmt.Println("error while changing proxy URL to target one:", err)
			return
		}

		/*reqID, err := p.requestRepo.AddRequest(req)
		if err != nil {
			fmt.Println("error while dumping the request:", err)
			return
		}*/

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("error while sending request to target:", err)
			return
		}
		defer response.Body.Close()

		/*_, err = p.responseRepo.AddResponse(response, reqID)
		if err != nil {
			fmt.Println("error while dumping the response:", err)
			return
		}*/

		err = response.Write(tlsConn)
		if err != nil {
			fmt.Println("error while sending response to the client:", err)
		}
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.httpsProxy(w, r)
	} else {
		p.httpProxy(w, r)
	}
}

func main() {
	addr := flag.String("addr", ":8080", "proxy address")
	flag.Parse()

	db, err := repository.GetPostgres()
	if err != nil {
		fmt.Println("error while initializing database:", err)
		return
	}

	requestRepo := repository.NewPsqlRequestRepository(db)
	responseRepo := repository.NewPsqlResponseRepository(db)

	proxy, err := createProxy(requestRepo, responseRepo)
	if err != nil {
		fmt.Println("error while creating proxy object:", err)
		return
	}

	fmt.Println("Starting proxy server on", *addr)
	err = http.ListenAndServe(*addr, proxy)
	if err != nil {
		fmt.Println("error while starting proxy:", err)
		return
	}
}
