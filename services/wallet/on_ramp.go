package wallet

import (
	"encoding/json"
	"errors"
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

type CryptoOnRampManager struct {
	dataSource string
	ramps      []CryptoOnRamp
	LastCalled time.Time
}

func NewCryptoOnRampManager(dataSource string) *CryptoOnRampManager {
	return &CryptoOnRampManager{
		dataSource: dataSource,
	}
}

// TODO(Samyoul) Make getting the datasource independent of http.Client
func (c *CryptoOnRampManager) Get() ([]CryptoOnRamp, error) {
	if c.dataSource == "" {
		return c.ramps, errors.New("data source is not set for CryptoOnRampManager")
	}

	if !c.hasCacheExpired(time.Now()) {
		return c.ramps, nil
	}

	sgc := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, c.dataSource, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "status-go")

	res, err := sgc.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &c.ramps)
	if err != nil {
		return nil, err
	}

	c.LastCalled = time.Now()
	return c.ramps, nil
}

func (c *CryptoOnRampManager) hasCacheExpired(t time.Time) bool {
	// If LastCalled + 1 hour is after the given time, then 1 hour hasn't passed yet
	if c.LastCalled.Add(time.Hour).After(t) {
		return false
	}

	return true
}
