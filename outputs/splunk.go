package outputs

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/tzurielweisberg/postee/v2/data"
	"github.com/tzurielweisberg/postee/v2/formatting"
	"github.com/tzurielweisberg/postee/v2/layout"
	"github.com/tzurielweisberg/postee/v2/log"
)

const (
	defaultSizeLimit = 10000
	SplunkType       = "splunk"
)

type SplunkOutput struct {
	Name         string
	Url          string
	Token        string
	EventLimit   int
	splunkLayout layout.LayoutProvider
}

func (splunk *SplunkOutput) GetType() string {
	return SplunkType
}

func (splunk *SplunkOutput) GetName() string {
	return splunk.Name
}

func (splunk *SplunkOutput) CloneSettings() *data.OutputSettings {
	return &data.OutputSettings{
		Name:      splunk.Name,
		Url:       splunk.Url,
		Token:     splunk.Token,
		SizeLimit: splunk.EventLimit,
		Enable:    true,
		Type:      SplunkType,
	}
}

func (splunk *SplunkOutput) Init() error {
	splunk.splunkLayout = new(formatting.HtmlProvider)
	log.Logger.Infof("Successfully initialized Splunk output %q", splunk.Name)
	return nil
}

func (splunk *SplunkOutput) Send(input map[string]string) (data.OutputResponse, error) {
	log.Logger.Infof("Sending to Splunk via %q", splunk.Name)

	if splunk.EventLimit == 0 {
		splunk.EventLimit = defaultSizeLimit
	}
	if splunk.EventLimit < defaultSizeLimit {
		log.Logger.Warnf("%q has a short limit %d (default %d)",
			splunk.Name, splunk.EventLimit, defaultSizeLimit)
	}

	if !strings.HasSuffix(splunk.Url, "/") {
		splunk.Url += "/"
	}

	rawEventData, ok := input["description"]
	if !ok {
		log.Logger.Error("Splunk sending error: empty content")
		return data.OutputResponse{}, nil
	}

	eventData := make(map[string]interface{})
	err := json.Unmarshal([]byte(rawEventData), &eventData)
	if err != nil {
		log.Logger.Errorf("sending to Splunk %q error: %v", splunk.Name, err)
		return data.OutputResponse{}, err
	}

	eventFormat := "{\"sourcetype\": \"_json\", \"event\": "
	constLimit := len(eventFormat) - 1

	_, imageFound := eventData["image"]

	var rawMsg []byte
	category, ok := eventData[EventCategoryAttribute]
	if ok && (category == CategoryIncident || category == CategoryInsights) {
		rawMsg = []byte(rawEventData)
	} else if resource, ok := eventData[ResourceTypeKey].(string); ok && resource == CodeRepoResource && !imageFound {
		rawMsg = []byte(rawEventData)
	} else {
		scanInfo := new(data.ScanImageInfo)
		err := json.Unmarshal([]byte(rawEventData), scanInfo)
		if err != nil {
			log.Logger.Errorf("sending to %q error: %v", splunk.Name, err)
			return data.OutputResponse{}, err
		}

		for {
			rawMsg, err = json.Marshal(scanInfo)
			if err != nil {
				log.Logger.Errorf("sending to Splunk %q error: %v", splunk.Name, err)
				return data.OutputResponse{}, err
			}
			if len(rawMsg) < splunk.EventLimit-constLimit {
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
					scanInfo.Image, splunk.Name, len(rawMsg), splunk.EventLimit)
				log.Logger.Infof(msg)
				return data.OutputResponse{}, errors.New(msg)
			}
		}
	}

	var buff bytes.Buffer
	buff.WriteString(eventFormat)
	buff.Write(rawMsg)
	buff.WriteByte('}')

	req, err := http.NewRequest("POST", splunk.Url+"services/collector", &buff)
	if err != nil {
		return data.OutputResponse{}, err
	}

	req.Header.Add("Authorization", "Splunk "+splunk.Token)

	client := http.Client{
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
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return data.OutputResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		log.Logger.Error(fmt.Errorf("splunk sending error: failed response status %q. Body: %q", resp.Status, string(b)))
		return data.OutputResponse{}, errors.New("failed response status for Splunk sending")
	}
	log.Logger.Infof("Sending a message to Splunk via %q was successful!", splunk.Name)
	return data.OutputResponse{}, nil
}

func (splunk *SplunkOutput) Terminate() error {
	log.Logger.Debugf("Splunk output %q terminated", splunk.Name)
	return nil
}

func (splunk *SplunkOutput) GetLayoutProvider() layout.LayoutProvider {
	return splunk.splunkLayout
}
