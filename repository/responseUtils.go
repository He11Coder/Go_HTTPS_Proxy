package repository

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type RawResponse struct {
	status      int
	jsonHeaders []byte
	jsonCookies []byte
	body        []byte
}

func parseResponse(resp *http.Response) (*RawResponse, error) {
	status := resp.StatusCode

	headers := map[string][]string{}
	for key, value := range resp.Header {
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
	cookies := resp.Cookies()
	if len(cookies) != 0 {
		jsonCookies, err = json.Marshal(cookies)
		if err != nil {
			return nil, err
		}
	}

	rawResp := &RawResponse{
		status:      status,
		jsonHeaders: jsonHeaders,
		jsonCookies: jsonCookies,
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := make([]byte, len(bodyBytes))
	copy(body, bodyBytes)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if len(body) != 0 {
		rawResp.body = body
	} else {
		rawResp.body = nil
	}

	return rawResp, nil
}
