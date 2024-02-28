package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type IRequestRepository interface {
	GetAllAPIRequests() ([]*APIRequest, error)
	GetAPIRequest(reqID int) (*APIRequest, error)
	GetRequest(reqID int) (*http.Request, error)
	GetRequestURL(reqID int) (*url.URL, error)
	AddRequest(req *http.Request) (int, error)
}

type psqlRequestRepository struct {
	requestStorage *sql.DB
}

func NewPsqlRequestRepository(db *sql.DB) IRequestRepository {
	return &psqlRequestRepository{
		requestStorage: db,
	}
}

func (p *psqlRequestRepository) GetAllAPIRequests() ([]*APIRequest, error) {
	rows, err := p.requestStorage.Query(`SELECT "id", "url", "method", "path", "query_param", "header", "cookie", "body" ` +
		`FROM request_data.request`)
	if errors.Is(sql.ErrNoRows, err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*APIRequest{}
	for rows.Next() {
		apiReq := &APIRequest{}
		var jsHeaders, jsCookies, body []byte

		err := rows.Scan(&apiReq.ID, &apiReq.URL, &apiReq.Method, &apiReq.Path, &apiReq.QueryParams, &jsHeaders, &jsCookies, &body)
		if err != nil {
			return nil, err
		}

		if jsHeaders != nil {
			err = json.Unmarshal(jsHeaders, &apiReq.Headers)
			if err != nil {
				return nil, err
			}
		} else {
			apiReq.Headers = nil
		}

		if jsCookies != nil {
			err = json.Unmarshal(jsCookies, &apiReq.Cookies)
			if err != nil {
				return nil, err
			}
		} else {
			apiReq.Cookies = nil
		}

		if body != nil {
			stringBody := string(body)
			apiReq.Body = &stringBody
		} else {
			apiReq.Body = nil
		}

		result = append(result, apiReq)
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func (p *psqlRequestRepository) GetAPIRequest(reqID int) (*APIRequest, error) {
	apiReq := &APIRequest{}
	var jsHeaders, jsCookies, body []byte

	err := p.requestStorage.QueryRow(`SELECT "id", "url", "method", "path", "query_param", "header", "cookie", "body" `+
		`FROM request_data.request WHERE id = $1`, reqID).
		Scan(&apiReq.ID, &apiReq.URL, &apiReq.Method, &apiReq.Path, &apiReq.QueryParams, &jsHeaders, &jsCookies, &body)
	if errors.Is(sql.ErrNoRows, err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if jsHeaders != nil {
		err = json.Unmarshal(jsHeaders, &apiReq.Headers)
		if err != nil {
			return nil, err
		}
	} else {
		apiReq.Headers = nil
	}

	if jsCookies != nil {
		err = json.Unmarshal(jsCookies, &apiReq.Cookies)
		if err != nil {
			return nil, err
		}
	} else {
		apiReq.Cookies = nil
	}

	if body != nil {
		stringBody := string(body)
		apiReq.Body = &stringBody
	} else {
		apiReq.Body = nil
	}

	return apiReq, nil
}

func (p *psqlRequestRepository) GetRequest(reqID int) (*http.Request, error) {
	rawReq := &RawRequest{}

	err := p.requestStorage.QueryRow(`SELECT "url", "method", "header", "cookie", "body" `+
		`FROM request_data.request WHERE id = $1`, reqID).Scan(&rawReq.url, &rawReq.method, &rawReq.jsonHeaders, &rawReq.jsonCookies, &rawReq.body)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	req, err := constructRequest(rawReq)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (p *psqlRequestRepository) GetRequestURL(reqID int) (*url.URL, error) {
	var reqURL string
	err := p.requestStorage.QueryRow(`SELECT "url" FROM request_data.request WHERE id = $1`, reqID).Scan(&reqURL)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return url.Parse(reqURL)
}

func (p *psqlRequestRepository) AddRequest(req *http.Request) (int, error) {
	rawReq, err := parseRequest(req)
	if err != nil {
		return 0, err
	}

	var requestID int
	err = p.requestStorage.QueryRow(`INSERT INTO request_data.request `+
		`("url", "method", "path", "query_param", "header", "cookie", "body") `+
		`VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		rawReq.url, rawReq.method, rawReq.path, rawReq.queryParams, rawReq.jsonHeaders, rawReq.jsonCookies, rawReq.body).Scan(&requestID)
	if err != nil {
		return 0, err
	}

	return requestID, nil
}
