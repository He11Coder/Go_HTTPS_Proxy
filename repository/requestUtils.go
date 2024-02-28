package repository

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type APIRequest struct {
	ID          int
	URL         string
	Method      string
	Path        *string
	QueryParams *string
	Headers     map[string][]string
	Cookies     []http.Cookie
	Body        *string
}

type RawRequest struct {
	url         string
	method      string
	path        *string
	queryParams *string
	jsonHeaders []byte
	jsonCookies []byte
	body        []byte
}

func parseRequest(req *http.Request) (*RawRequest, error) {
	method := req.Method
	path := req.URL.Path
	queryParams := req.URL.Query().Encode()
	url := req.URL.String()

	headers := map[string][]string{}
	for key, value := range req.Header {
		headers[key] = value
	}

	var jsonHeaders []byte = nil
	var err error
	if len(headers) != 0 {
		jsonHeaders, err = json.Marshal(headers)
		if err != nil {
			return nil, err
		}
	}

	var jsonCookies []byte = nil
	cookies := req.Cookies()
	if len(cookies) != 0 {
		jsonCookies, err = json.Marshal(cookies)
		if err != nil {
			return nil, err
		}
	}

	rawReq := &RawRequest{
		url:         url,
		method:      method,
		jsonHeaders: jsonHeaders,
		jsonCookies: jsonCookies,
	}

	if path == "" {
		rawReq.path = nil
	} else {
		rawReq.path = &path
	}

	if queryParams == "" {
		rawReq.queryParams = nil
	} else {
		rawReq.queryParams = &queryParams
	}

	defer req.Body.Close()
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	body := make([]byte, len(bodyBytes))
	copy(body, bodyBytes)
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if len(body) != 0 {
		rawReq.body = body
	} else {
		rawReq.body = nil
	}

	/*if req.Body != nil {
		bodyReader, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		defer bodyReader.Close()

		body, err := io.ReadAll(bodyReader)
		if err != nil {
			return nil, err
		}

		rawReq.body = body
	} else {
		rawReq.body = nil
	}*/

	return rawReq, nil
}

func constructRequest(raw *RawRequest) (*http.Request, error) {
	b := bytes.NewReader(raw.body)

	req, err := http.NewRequest(raw.method, raw.url, b)
	if err != nil {
		return nil, err
	}

	headerMap := map[string][]string{}
	err = json.Unmarshal(raw.jsonHeaders, &headerMap)
	if err != nil {
		return nil, err
	}

	for key, values := range headerMap {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	/*cookies := []http.Cookie{}
	json.Unmarshal(raw.jsonCookies, &cookies)

	for _, cookie := range cookies {
		req.AddCookie(&cookie)
	}*/

	return req, nil
}
