package teams_api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tzurielweisberg/postee/v2/log"
)

func CreateMessageByWebhook(webhook, content string) error {
	var message bytes.Buffer
	fmt.Fprintf(&message, "{\"text\":\"%s\"}", content)

	log.Logger.Debugf("Data for sending to %q: %q", webhook, message.String())
	r := bytes.NewReader(message.Bytes())
	client := http.DefaultClient
	reg, err := http.NewRequest("POST", webhook, r)
	if err != nil {
		return err
	}
	reg.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(reg)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if message, _ := ioutil.ReadAll(resp.Body); resp.StatusCode != http.StatusOK {
		return fmt.Errorf("InsertRecordToTable Error: %q. %s", resp.Status, message)
	} else {
		log.Logger.Debugf("Response body: %q", message)
	}
	return nil
}
