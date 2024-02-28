package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"proxy/repository"
	"proxy/scanner"
)

var respDataTmpl = `<table border="1">
<tr>
	<th>Request ID</th>
	<th>URL</th>
	<th>HTTP Method</th>
	<th>Path</th>
	<th>Query Parameters</th>
	<th>Headers</th>
	<th>Cookies</th>
	<th>Body</th>
</tr>
{{range .}}
<tr>
	<td>{{.ID}}</td>
	<td>{{.URL}}</td>
	<td>{{.Method}}</td>
	<td>{{.Path}}</td>
	<td>{{.QueryParams}}</td>
	<td>
		<ul>
			{{range $key, $value := .Headers}}  
				<li><strong>{{$key}}:</strong> {{$value}}</li>
			{{end}}
		</ul>
	</td>
	<td>
		<ul>
			{{range $_, $cookie := .Cookies}}
				<li><strong>Name:</strong> {{$cookie.Name}}</li>
				<li><strong>Value:</strong> {{$cookie.Value}}</li>
				<li><strong>Path:</strong> {{$cookie.Path}}</li>
				<li><strong>Domain:</strong> {{$cookie.Domain}}</li>
				<li><strong>Expires:</strong> {{$cookie.Expires}}</li>
				<li><strong>Secure:</strong> {{$cookie.Secure}}</li>
				<li><strong>HttpOnly:</strong> {{$cookie.HttpOnly}}</li>
				<br>
			{{end}}
		</ul>
	</td>
	<td>{{.Body}}</td>
</tr>
{{end}}
</table>`

var scanTmpl = `<strong>The following vulnerable paths have been found for this resource:</strong>
<ul>
	{{range .}}
		<li>
			{{.URL}} - <strong>{{.ResponseCode}}</strong> <br><br> <strong>Body:</strong> <br>
			{{.Body | html}}
		</li>
		<hr>
	{{end}}
</ul>`

type ProxyAPIHandler struct {
	requestRepository repository.IRequestRepository
}

func NewAPIHandler(reqRepo repository.IRequestRepository) *ProxyAPIHandler {
	return &ProxyAPIHandler{
		requestRepository: reqRepo,
	}
}

func (h *ProxyAPIHandler) getAllRequests(w http.ResponseWriter, r *http.Request) {
	reqs, err := h.requestRepository.GetAllAPIRequests()
	if reqs == nil && err == nil {
		http.Error(w, "no requests found", http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("error while getting all requests:", err)
		http.Error(w, "the server could not process your request", http.StatusInternalServerError)
		return
	}

	t := template.Must(template.New("table").Parse(respDataTmpl))
	t.Execute(w, reqs)
}

func (h *ProxyAPIHandler) getRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqID, err := strconv.Atoi(vars["reqID"])
	if err != nil {
		http.Error(w, "request id must be integer", http.StatusBadRequest)
		return
	}

	req, err := h.requestRepository.GetAPIRequest(reqID)
	if req == nil && err == nil {
		http.Error(w, "no requests found", http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("error while getting request by id:", err)
		http.Error(w, "the server could not process your request", http.StatusInternalServerError)
		return
	}

	t := template.Must(template.New("table").Parse(respDataTmpl))
	t.Execute(w, []*repository.APIRequest{req})
}

func (h *ProxyAPIHandler) repeatRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqID, err := strconv.Atoi(vars["reqID"])
	if err != nil {
		http.Error(w, "request id must be integer", http.StatusBadRequest)
		return
	}

	req, err := h.requestRepository.GetRequest(reqID)
	if req == nil && err == nil {
		http.Error(w, "no requests found", http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("error while getting request by id to repeat it:", err)
		http.Error(w, "the server could not process your request", http.StatusInternalServerError)
		return
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error while repeating request:", err)
		http.Error(w, "the server could not repeat your request", http.StatusBadRequest)
		return
	}

	response.Write(w)
}

func (h *ProxyAPIHandler) scanRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqID, err := strconv.Atoi(vars["reqID"])
	if err != nil {
		http.Error(w, "request id must be integer", http.StatusBadRequest)
		return
	}

	reqURL, err := h.requestRepository.GetRequestURL(reqID)
	if reqURL == nil && err == nil {
		http.Error(w, "no requests found", http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("error while getting request URL by id to scan it:", err)
		http.Error(w, "invalid url to scan", http.StatusBadRequest)
		return
	}

	vulScanner := scanner.NewScanner(reqURL)
	vulnerableURLs, err := vulScanner.Dirbust()
	if err != nil {
		fmt.Println("error while dirbusting the URL:", err)
		http.Error(w, "dirbusting error", http.StatusInternalServerError)
		return
	}

	if len(vulnerableURLs) != 0 {
		t := template.Must(template.New("scanResult").Parse(scanTmpl))

		w.Header().Add("Content-Type", "text/html")
		t.Execute(w, vulnerableURLs)
	} else {
		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte("<strong>No vulnerabilities found!</strong>"))
	}
}

func main() {
	addr := flag.String("addr", ":8000", "proxy API address")
	flag.Parse()

	db, err := repository.GetPostgres()
	if err != nil {
		fmt.Println("error while connecting to database (proxy API):", err)
		return
	}
	defer db.Close()

	reqRepo := repository.NewPsqlRequestRepository(db)
	handler := NewAPIHandler(reqRepo)

	router := mux.NewRouter()

	router.HandleFunc("/requests", handler.getAllRequests).Methods("GET")
	router.HandleFunc("/requests/{reqID}", handler.getRequest).Methods("GET")
	router.HandleFunc("/repeat/{reqID}", handler.repeatRequest).Methods("GET")
	router.HandleFunc("/scan/{reqID}", handler.scanRequest).Methods("GET")

	http.Handle("/", router)

	fmt.Println("Starting proxy server API on", *addr)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Println("error while starting proxy API:", err)
		return
	}
}
