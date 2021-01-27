package wallet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	cryptoOnRampsData = "https://raw.githubusercontent.com/status-im/crypto-on-ramps/master/ramps.json"

	DataSourceHTTP DataSourceType = iota + 1
	DataSourceStatic
)

type CryptoOnRamp struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Fees        string `json:"fees"`
	Region      string `json:"region"`
	LogoURL     string `json:"logoUrl"`
	SiteURL     string `json:"siteUrl"`
}

type DataSourceType int

type CryptoOnRampOptions struct {
	dataSource     string
	dataSourceType DataSourceType
}

type CryptoOnRampManager struct {
	options    *CryptoOnRampOptions
	ramps      []CryptoOnRamp
	LastCalled time.Time
}

func NewCryptoOnRampManager(options *CryptoOnRampOptions) *CryptoOnRampManager {
	return &CryptoOnRampManager{
		options: options,
	}
}

func (c *CryptoOnRampManager) Get() ([]CryptoOnRamp, error) {
	if !c.hasCacheExpired(time.Now()) {
		return c.ramps, nil
	}

	var data []byte
	var err error

	switch c.options.dataSourceType {
	case DataSourceHTTP:
		data, err = c.getFromHttpDataSource()
	case DataSourceStatic:
		data, err = c.getFromStaticDataSource()
	default:
		return nil, fmt.Errorf("unsupported CryptoOnRampManager.dataSourceType '%d'", c.options.dataSourceType)
	}
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &c.ramps)
	if err != nil {
		return nil, err
	}

	return c.ramps, nil
}

func (c *CryptoOnRampManager) hasCacheExpired(t time.Time) bool {
	// If LastCalled + 1 hour is after the given time, then 1 hour hasn't passed yet
	if c.LastCalled.Add(time.Hour).After(t) {
		return false
	}

	return true
}

func (c *CryptoOnRampManager) getFromHttpDataSource() ([]byte, error) {
	if c.options.dataSource == "" {
		return nil, errors.New("data source is not set for CryptoOnRampManager")
	}

	sgc := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest(http.MethodGet, c.options.dataSource, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "status-go")

	res, err := sgc.Do(req)
	if err != nil {
		return nil, err
	}

	c.LastCalled = time.Now()

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *CryptoOnRampManager) getFromStaticDataSource() ([]byte, error) {
	data := `
[
  {
    "name": "Wyre",
    "description": "A secure bridge for fiat and crypto",
    "fees": "from 2.9%",
    "region": "US & Europe",
    "logo-url":"https://www.sendwyre.com/favicon.ico",
    "site-url": "https://pay.sendwyre.com/purchase"
  },
  {
    "name": "MoonPay",
    "description": "The new standard for fiat to crypto",
    "fees": "1%-4.5%",
    "region": "US & Europe",
    "logo-url":"https://buy.moonpay.com/favicon-32x32.png",
    "site-url": "https://buy.moonpay.com"
  },
  {
    "name": "Transak",
    "description": "Global fiat <-> crypto payment gateway",
    "fees": "1%-4.5%",
    "region": "Global",
    "logo-url":"https://global.transak.com/favicon.png",
    "site-url": "https://global.transak.com"
  },
  {
    "name": "Ramp",
    "description": "Global crypto to fiat flow",
    "fees": "1.5%",
    "region": "Global",
    "logo-url":"https://ramp.network/assets/favicons/favicon-32x32.png",
    "site-url": "https://ramp.network/buy/"
  },
  {
    "name": "LocalCryptos",
    "description": "Non-custodial crypto marketplace",
    "fees": "1.5%",
    "region": "Global",
    "logo-url":"https://localcryptos.com/images/favicon.png",
    "site-url": "https://localcryptos.com"
  }
]`

	return []byte(data), nil
}
