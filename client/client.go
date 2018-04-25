// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

/*
	Package agent implements the command line interface to interact with the
	API rest
*/
package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"verifiabledata/api/apihttp"
	"verifiabledata/balloon"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
)

// TODO: write documentation
type Client interface {
	InsertEvent([]byte) *apihttp.Snapshot
	Membership([]byte, uint64) *apihttp.MembershipProof
	Verify([]byte, *balloon.Commitment, *apihttp.MembershipProof) bool
}

type HttpClient struct {
	httpEndpoint string
	apiKey       string
	client       *http.Client
}

func NewHttpClient(endpoint, apiKey string, c *http.Client) *HttpClient {
	return &HttpClient{
		endpoint,
		key,
		c,
	}
}

func (c Client) InsertEvent(event []byte) apihttp.Snapshot {

	// TODO: rename message to event also in apiHttp
	data := []byte(strings.Join([]string{`{"message": "`, event, `"}`}, ""))

	req, err := http.NewRequest("POST", c.httpEndpoint+"/events", bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var snapshot apihttp.Snapshot

	json.Unmarshal([]byte(bodyBytes), &snapshot)

	return &snapshot

}

func (c Client) Membership(event []byte, version uint64) *apihttp.MembershipProof {

	req, err := http.NewRequest("GET", c.httpEndpoint+"/proofs/membership", nil)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Set("key", string(version))
	q.Set("version", string(version))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var proof apihttp.MembershipProof

	json.Unmarshal([]byte(bodyBytes), &proof)

	return &proof

}

func (c Client) Verify(event []byte, cm *balloon.Commitment, p *apihttp.MembershipProof) bool {
	htlh := history.LeafHasherF(hashing.Sha256Hasher)
	htih := history.InteriorHasherF(hashing.Sha256Hasher)

	hylh := hyper.LeafHasherF(hashing.Sha256Hasher)
	hyih := hyper.InteriorHasherF(hashing.Sha256Hasher)

	historyProof := history.NewProof(p.Proofs.HistoryAuditPath, htlf, htih)
	hyperProof := hyper.NewProof(p.Proofs.HyperAuditPath, hylf, hyih)

	proof := balloon.NewProof(p.IsMember, hyperProof, historyProof, p.QueryVersion, p.ActualVersion)

	return proof.Verify(cm, event)
}
