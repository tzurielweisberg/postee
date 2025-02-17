package actions

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tzurielweisberg/postee/v2/formatting"
	"github.com/tzurielweisberg/postee/v2/layout"
)

type WebhookAction struct {
	Name    string
	Url     string
	Timeout string
}

func (webhook *WebhookAction) GetName() string {
	return webhook.Name
}

func (webhook *WebhookAction) Init() error {
	log.Printf("Starting Webhook action %q, for sending to %q",
		webhook.Name, webhook.Url)
	return nil
}

func (webhook *WebhookAction) Send(content map[string]string) error {
	log.Printf("Sending webhook to %q", webhook.Url)
	data := content["description"] //it's not supposed to work with legacy renderer
	client, err := newClient(webhook.Timeout)
	if err != nil {
		return err
	}
	resp, err := client.Post(webhook.Url, "application/json", strings.NewReader(data))
	if err != nil {
		log.Printf("Sending webhook Error: %v", err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Sending %q Error: %v", webhook.Name, err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		msg := "Sending webhook wrong status: '%d'. Body: %s"
		log.Printf(msg, resp.StatusCode, body)
		return fmt.Errorf(msg, resp.StatusCode, body)
	}
	log.Printf("Sending Webhook to %q was successful!", webhook.Name)
	return nil
}

func (webhook *WebhookAction) Terminate() error {
	log.Printf("Webhook action %q terminated.", webhook.Name)
	return nil
}

func (webhook *WebhookAction) GetLayoutProvider() layout.LayoutProvider {
	// Todo: This is MOCK. Because Formatting isn't need for Webhook
	// todo: The App should work with `return nil`
	return new(formatting.HtmlProvider)
}

var newClient = func(timeout string) (http.Client, error) {
	if len(timeout) == 0 || timeout == "0" {
		timeout = "120s"
	}
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		return http.Client{}, fmt.Errorf("invalid duration specified: %w", err)
	}
	return http.Client{Timeout: duration}, nil
}
