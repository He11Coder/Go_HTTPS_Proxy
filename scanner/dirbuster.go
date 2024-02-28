package scanner

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ResourceInfo struct {
	URL          string
	ResponseCode int
	Body         *string
}

type Scanner struct {
	URLToScan *url.URL
}

func NewScanner(url *url.URL) *Scanner {
	return &Scanner{
		URLToScan: url,
	}
}

func (sc *Scanner) Dirbust() ([]*ResourceInfo, error) {
	host := strings.TrimSuffix(sc.URLToScan.String(), sc.URLToScan.Path)

	vulPaths, err := os.Open("./api/dicc.txt")
	if err != nil {
		return nil, err
	}
	defer vulPaths.Close()

	fileScanner := bufio.NewScanner(vulPaths)

	client := http.Client{}
	i := 1
	vulURLs := []*ResourceInfo{}
	for fileScanner.Scan() && i < 500 {
		vulURL := host + "/" + fileScanner.Text()

		req, err := http.NewRequest("GET", vulURL, nil)
		if err != nil {
			fmt.Println("error while constructing vulnerable request:", err)
			continue
		}

		response, err := client.Do(req)
		if err != nil {
			fmt.Println("error while sending vulnerable request:", err)
			continue
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusNotFound {
			vulRes := &ResourceInfo{
				URL:          vulURL,
				ResponseCode: response.StatusCode,
			}

			b, err := io.ReadAll(response.Body)
			if err != nil {
				mes := "Can not read response body"
				vulRes.Body = &mes
			} else if len(b) == 0 {
				vulRes.Body = nil
			} else {
				body := string(b)
				vulRes.Body = &body
			}

			vulURLs = append(vulURLs, vulRes)
		}
		i++
	}

	return vulURLs, nil
}
