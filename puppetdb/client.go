package puppetdb

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type PuppetDBClient struct {
	URL  string
	Cert string
	Key  string
	CA   string
}

type PuppetDBResp struct {
	Error                        error  `json:"error"`
	Certname                     string `json:"certname"`
	Deactivated                  string `json:"deactivated"`
	Expired                      string `json:"expired"`
	CachedCatalogStatus          string `json:"cached_catalog_status"`
	CatalogEnvironment           string `json:"catalog_environment"`
	FactsEnvironment             string `json:"facts_environment"`
	ReportEnvironment            string `json:"report_environment"`
	CatalogTimestamp             string `json:"catalog_timestamp"`
	FactsTimestamp               string `json:"facts_timestamp"`
	ReportTimestamp              string `json:"report_timestamp"`
	LatestReportCorrectiveChange string `json:"latest_report_corrective_change"`
	LatestReportHash             string `json:"latest_report_hash"`
	LatestReportNoop             bool   `json:"latest_report_noop"`
	LatestReportNoopPending      bool   `json:"latest_report_noop_pending"`
	LatestReportStatus           string `json:"latest_report_status"`
}

type commandsPayload struct {
	Command string            `json:"command"`
	Version int               `json:"version"`
	Payload map[string]string `json:"payload"`
}

func isFile(str string) bool {
	return strings.HasPrefix(str, "/")
}

func (p *PuppetDBClient) Query(query string, verb string, payload string) (pdbResp PuppetDBResp, err error) {
	url := p.URL + "/pdb/" + query
	form := strings.NewReader(payload)
	req, err := http.NewRequest(verb, url, form)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	var client *http.Client
	if p.Cert != "" {
		// Load cert pair
		var cert tls.Certificate
		if isFile(p.Cert) {
			if !isFile(p.Key) {
				err = fmt.Errorf("cert points to a file but key is a string: %s", err)
				return
			}

			cert, err = tls.LoadX509KeyPair(p.Cert, p.Key)
			if err != nil {
				return pdbResp, err
			}
		} else {
			if isFile(p.Key) {
				err = fmt.Errorf("cert is a string but key points to a file: %s", err)
				return
			}

			cert, err = tls.X509KeyPair([]byte(p.Cert), []byte(p.Key))
			if err != nil {
				err = fmt.Errorf("failed to load client cert from string: %s", err)
				return
			}
		}

		// Load CA cert
		var caCert []byte
		if isFile(p.CA) {
			caCert, err = ioutil.ReadFile(p.CA)
			if err != nil {
				return pdbResp, fmt.Errorf("failed to load CA cert at %s: %s", p.CA, err)
			}
		} else {
			caCert = []byte(p.CA)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	json.Unmarshal(body, &pdbResp)

	if err = pdbResp.Error; err != nil {
		return
	}

	return pdbResp, nil
}
