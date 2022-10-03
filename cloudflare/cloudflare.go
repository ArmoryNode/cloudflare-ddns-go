package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/ArmoryNode/cloudflare-ddns/cli"
)

type CloudflareClient struct {
	ApiToken                string `json:"apiToken"`
	ZoneIdentifier          string `json:"zoneIdentifier"`
	RecordIdentifier        string `json:"recordIdentifier"`
	UpdateIntervalInMinutes int    `json:"updateIntervalInMinutes"`

	httpClient *http.Client
}

type CloudflareResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
}

var currentIPv4 = ""
var lastUpdated = time.Now().Format(time.RFC1123)

var UpdateInterval int

func CreateClient(apiToken string, zoneIdentifier string, recordIdentifier string, updateInterval int) *CloudflareClient {
	return &CloudflareClient{
		ApiToken:                apiToken,
		ZoneIdentifier:          zoneIdentifier,
		RecordIdentifier:        recordIdentifier,
		UpdateIntervalInMinutes: updateInterval,
		httpClient:              &http.Client{},
	}
}

func (cc CloudflareClient) createRequest(method string, url string, data io.Reader) *http.Request {
	token := "Bearer\u0020" + cc.ApiToken
	req, err := http.NewRequest(method, url, data)

	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Authorization", token)

	return req
}

func (cc CloudflareClient) VerifyApiToken(done chan bool) bool {
	req := cc.createRequest(http.MethodGet, "https://api.cloudflare.com/client/v4/user/tokens/verify", nil)

	resp, err := cc.httpClient.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var response CloudflareResponse

	json.NewDecoder(resp.Body).Decode(&response)

	done <- true
	return response.Success
}

func (cc CloudflareClient) VerifyZone(done chan bool) bool {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s", cc.ZoneIdentifier)

	client := &http.Client{}
	req := cc.createRequest(http.MethodGet, url, nil)

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var response CloudflareResponse

	json.NewDecoder(resp.Body).Decode(&response)

	done <- true
	return response.Success
}

func (cc CloudflareClient) VerifyDnsRecord(done chan bool) bool {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", cc.ZoneIdentifier, cc.RecordIdentifier)

	client := &http.Client{}
	req := cc.createRequest(http.MethodGet, url, nil)

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var response CloudflareResponse

	json.NewDecoder(resp.Body).Decode(&response)

	done <- true
	return response.Success
}

func (cc CloudflareClient) UpdateRecord() {
	ipv4 := getPublicIPv4()
	client := &http.Client{}

	if ipv4 == currentIPv4 {
		fmt.Printf("\n%s\rIP Address unchanged. Last updated: %s", cli.CLEAR_LINE, lastUpdated)
		fmt.Print(cli.UP_LINE)
		return
	}

	data, err := json.Marshal(map[string]interface{}{
		"content": ipv4,
	})

	if err != nil {
		log.Fatalln(err)
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", cc.ZoneIdentifier, cc.RecordIdentifier)

	request := cc.createRequest(http.MethodPatch, url, bytes.NewBuffer(data))

	resp, err := client.Do(request)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		currentIPv4 = ipv4
		lastUpdated = time.Now().Format(time.RFC1123)
		fmt.Printf("\n\rLast Updated: %s", lastUpdated)
		fmt.Print(cli.UP_LINE)
	} else {
		fmt.Printf("Failed to update DNS record (status code: %d)", resp.StatusCode)
		b, _ := io.ReadAll(resp.Body)
		fmt.Print(string(b))
	}
}

func getPublicIPv4() string {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "https://ifconfig.co/ip", nil)

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	r := regexp.MustCompile(`((?:[0-9]{1,3}\.){3}[0-9]{1,3})`)

	// ifconfig.co returns an IP address with a newline character at the end, so we're using regex to chop it off
	return r.FindString(string(bytes))
}
