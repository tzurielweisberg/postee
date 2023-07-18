package actions

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/tzurielweisberg/postee/v2/data"
	"github.com/tzurielweisberg/postee/v2/formatting"
	"github.com/tzurielweisberg/postee/v2/layout"
)

const defaultSizeLimit = 10000

type SplunkAction struct {
	Name         string
	Url          string
	Token        string
	EventLimit   int
	TlsVerify    bool
	splunkLayout layout.LayoutProvider
}

func (splunk *SplunkAction) GetName() string {
	return splunk.Name
}

func (splunk *SplunkAction) Init() error {
	splunk.splunkLayout = new(formatting.HtmlProvider)
	log.Printf("Starting Splunk action %q....", splunk.Name)
	return nil
}

func (splunk *SplunkAction) Send(d map[string]string) error {
	log.Printf("Sending a message to %q", splunk.Name)

	if splunk.EventLimit == 0 {
		splunk.EventLimit = defaultSizeLimit
	}
	if splunk.EventLimit < defaultSizeLimit {
		log.Printf("[WARNING] %q has a short limit %d (default %d)",
			splunk.Name, splunk.EventLimit, defaultSizeLimit)
	}

	if !strings.HasSuffix(splunk.Url, "/") {
		splunk.Url += "/"
	}

	scanInfo := new(data.ScanImageInfo)
	body := []byte(d["description"])
	if !json.Valid([]byte(body)) {
		return errors.New("wrong template selected, choose a correct template")
	}

	err := json.Unmarshal(body, scanInfo)
	if err != nil {
		log.Printf("sending to %q error: %v", splunk.Name, err)
		return err
	}

	eventFormat := "{\"sourcetype\": \"_json\", \"event\": "
	constLimit := len(eventFormat) - 1

	var fields []byte

	for {
		fields, err = json.Marshal(scanInfo)
		if err != nil {
			log.Printf("sending to %q error: %v", splunk.Name, err)
			return err
		}
		if len(fields) < splunk.EventLimit-constLimit {
			break
		}
		switch {
		case len(scanInfo.Resources) > 0:
			scanInfo.Resources = nil
			continue
		case len(scanInfo.Malwares) > 0:
			scanInfo.Malwares = nil
			continue
		case len(scanInfo.SensitiveData) > 0:
			scanInfo.SensitiveData = nil
			continue
		default:
			msg := fmt.Sprintf("Scan result for %q is large for %q , its size if %d (limit %d)",
				scanInfo.Image, splunk.Name, len(fields), splunk.EventLimit)
			log.Print(msg)
			return errors.New(msg)
		}
	}

	var buff bytes.Buffer
	buff.WriteString(eventFormat)
	buff.Write(fields)
	buff.WriteByte('}')

	req, err := http.NewRequest("POST", splunk.Url+"services/collector", &buff)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Splunk "+splunk.Token)

	client := http.Client{
		// default transport with tls config added
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: splunk.TlsVerify},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Splunk sending error: failed response status %q. Body: %q", resp.Status, string(b))
		return errors.New("failed response status for Splunk sending")
	}
	log.Printf("Sending a message to %q was successful!", splunk.Name)
	return nil
}

func (splunk *SplunkAction) Terminate() error {
	log.Printf("Splunk action %q terminated", splunk.Name)
	return nil
}

func (splunk *SplunkAction) GetLayoutProvider() layout.LayoutProvider {
	return splunk.splunkLayout
}
