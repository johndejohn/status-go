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

type CryptoOnRampManager struct {
	ramps []CryptoOnRamp
	LastCalled time.Time
}

func (c *CryptoOnRampManager) Get() ([]CryptoOnRamp, error) {
	// TODO deal with case where the c.LastCalled is not yet set
	if c.LastCalled.Add(time.Hour).Before(time.Now()) {
		return c.ramps, nil
	}

	sgc := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, cryptoOnRampsData, nil)
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
