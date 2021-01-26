package wallet

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	cryptoOnRampsData = "https://raw.githubusercontent.com/status-im/crypto-on-ramps/master/ramps.json"
)

type CryptoOnRamp struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Fees        string `json:"fees"`
	Region      string `json:"region"`
	LogoURL     string `json:"logoUrl"`
	SiteURL     string `json:"siteUrl"`
}

type CryptoOnRamps []CryptoOnRamp

// TODO introduce some caching to cache results and prevent spamming, maybe an hour's cache
func (c *CryptoOnRamps) Get() error {
	sgc := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, cryptoOnRampsData, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "status-go")

	res, err := sgc.Do(req)
	if err != nil {
		return err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &c)
	if err != nil {
		return err
	}

	return nil
}
