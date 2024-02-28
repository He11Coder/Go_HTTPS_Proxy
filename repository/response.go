package repository

import (
	"database/sql"
	"errors"
	"net/http"
)

type IResponseRepository interface {
	GetResponse(respID int) (*RawResponse, error)
	GetResponseByRequestID(reqID int) (*RawResponse, error)
	AddResponse(resp *http.Response, reqID int) (int, error)
}

type psqlResponseRepository struct {
	responseStorage *sql.DB
}

func NewPsqlResponseRepository(db *sql.DB) IResponseRepository {
	return &psqlResponseRepository{
		responseStorage: db,
	}
}

func (p *psqlResponseRepository) GetResponse(respID int) (*RawResponse, error) {
	rawResp := &RawResponse{}

	err := p.responseStorage.QueryRow(`SELECT "status_code", "header", "cookie", "body" `+
		`FROM request_data.response WHERE id = $1`, respID).Scan(&rawResp.status, &rawResp.jsonHeaders, &rawResp.jsonCookies, &rawResp.body)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return rawResp, nil
}

func (p *psqlResponseRepository) GetResponseByRequestID(reqID int) (*RawResponse, error) {
	rawResp := &RawResponse{}

	err := p.responseStorage.QueryRow(`SELECT status_code, header, cookie, body `+
		`FROM request_data.response WHERE id = (SELECT response_id FROM request_data.request WHERE id = $1)`, reqID).
		Scan(&rawResp.status, &rawResp.jsonHeaders, &rawResp.jsonCookies, &rawResp.body)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, nil
	}

	return rawResp, nil
}

func (p *psqlResponseRepository) AddResponse(resp *http.Response, reqID int) (int, error) {
	rawResp, err := parseResponse(resp)
	if err != nil {
		return 0, err
	}

	tx, err := p.responseStorage.Begin()
	if err != nil {
		return 0, err
	}

	var responseID int
	err = tx.QueryRow(`INSERT INTO request_data.response `+
		`("status_code", "header", "cookie", "body") `+
		`VALUES ($1, $2, $3, $4) RETURNING id`,
		rawResp.status, rawResp.jsonHeaders, rawResp.jsonCookies, rawResp.body).Scan(&responseID)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return 0, nil
		}
		return 0, err
	}

	_, err = tx.Exec(`UPDATE request_data.request SET response_id = $1 WHERE id = $2`, responseID, reqID)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return 0, rollbackErr
		}
		return 0, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return 0, commitErr
	}

	return responseID, nil
}
