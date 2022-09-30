package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	cli "github.com/armorynode/cloudflare-ddns/util"
)

// ASCII control characters
const UP_LINE = "\033[A"
const HIDE_CURSOR = "\033[?25l"
const CLEAR_LINE = "\x1b[2K"

// Used as an indicator that the application is running
var spinner = [4]rune{'|', '\\', '-', '/'}

var currentIPv4 = ""
var lastUpdated = time.Now().Format(time.RFC1123)

func main() {
	// Hide the cursor in the console
	fmt.Print(HIDE_CURSOR)

	var config CloudflareConfiguration
	getCloudflareConfiguration(&config)

	// Channel used to stop displaying the loading spinner
	quit := make(chan bool)

	// Channel used for checking the cloudflare configuration
	verified := make(chan bool)

	go cli.DisplaySpinner("Verifying cloudflare API token", quit)
	go verifyApiToken(verified, quit, config)

	if <-verified {
		fmt.Print(CLEAR_LINE)
		fmt.Println("[✓] API token verified")
	} else {
		fmt.Print(CLEAR_LINE)
		fmt.Println("[✕] Failed to verify API token")
		return
	}

	go cli.DisplaySpinner("Verifying zone", quit)
	go verifyZone(verified, quit, config)

	if <-verified {
		fmt.Print(CLEAR_LINE)
		fmt.Println("[✓] Zone verified")
	} else {
		fmt.Print(CLEAR_LINE)
		fmt.Println("[✕] Failed to verify zone")
		return
	}

	go cli.DisplaySpinner("Verifying DNS record", quit)
	go verifyDnsRecord(verified, quit, config)

	if <-verified {
		fmt.Print(CLEAR_LINE)
		fmt.Println("[✓] DNS record verified")
	} else {
		fmt.Print(CLEAR_LINE)
		fmt.Println("[✕] Failed to verify DNS record")
		return
	}

	// We won't be using these channels anymore
	close(verified)
	close(quit)

	// Pad the countdown down in the console
	fmt.Println()

	// Channel used for signaling the program to send an update request to cloudflare
	update := make(chan bool)

	//  We want to send an initial update to cloudflare, regardless of the current IP
	updateRecord(config)

	// Begin countdown
	go countdown(config.UpdateIntervalInMinutes, update)

	for {
		// Once an update signal is received, we will send a request to cloudflare
		<-update
		updateRecord(config)
	}
}

type CloudflareConfiguration struct {
	ApiToken                string `json:"apiToken"`
	ZoneIdentifier          string `json:"zoneIdentifier"`
	RecordIdentifier        string `json:"recordIdentifier"`
	UpdateIntervalInMinutes int    `json:"updateIntervalInMinutes"`
}

func getCloudflareConfiguration(config *CloudflareConfiguration) {
	jsonFile, err := os.Open("./config.json")
	if err != nil {
		log.Fatalln(err)
	}

	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	json.Unmarshal(byteValue, config)
}

type CloudflareVerificationResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
}

func verifyApiToken(verified chan bool, quit chan bool, config CloudflareConfiguration) {
	client := &http.Client{
		CheckRedirect: http.DefaultClient.CheckRedirect,
	}

	token := fmt.Sprintf("Bearer %s", config.ApiToken)
	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/user/tokens/verify", nil)
	req.Header.Add("Authorization", token)

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var response CloudflareVerificationResponse

	json.NewDecoder(resp.Body).Decode(&response)

	quit <- true

	if response.Success {
		verified <- true
	} else {
		verified <- false
	}
}

func verifyZone(verified chan bool, quit chan bool, config CloudflareConfiguration) {
	client := &http.Client{}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s", config.ZoneIdentifier)
	token := fmt.Sprintf("Bearer %s", config.ApiToken)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", token)

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var response CloudflareVerificationResponse

	json.NewDecoder(resp.Body).Decode(&response)

	quit <- true

	if response.Success {
		verified <- true
	} else {
		verified <- false
	}
}

func verifyDnsRecord(verified chan bool, quit chan bool, config CloudflareConfiguration) {
	client := &http.Client{}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", config.ZoneIdentifier, config.RecordIdentifier)
	token := fmt.Sprintf("Bearer %s", config.ApiToken)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", token)

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var response CloudflareVerificationResponse

	json.NewDecoder(resp.Body).Decode(&response)

	quit <- true

	if response.Success {
		verified <- true
	} else {
		verified <- false
	}
}

func countdown(minutes int, update chan bool) {
	secondsFrom := minutes * 60
	for {
		if secondsFrom >= 0 {
			fmt.Print(CLEAR_LINE)
			fmt.Printf("\r[%c] Updating in %s", spinner[secondsFrom%4], getTimeFromSeconds(secondsFrom))
			secondsFrom--
			time.Sleep(time.Second)
		} else {
			secondsFrom = minutes * 60
			update <- true
		}
	}
}

func updateRecord(config CloudflareConfiguration) {
	ipv4 := getPublicIPv4()

	if ipv4 == currentIPv4 {
		fmt.Printf("\n%s\rIP Address unchanged. Last updated: %s", CLEAR_LINE, lastUpdated)
		fmt.Print(UP_LINE)
		return
	}

	client := &http.Client{
		CheckRedirect: http.DefaultClient.CheckRedirect,
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", config.ZoneIdentifier, config.RecordIdentifier)
	token := fmt.Sprintf("Bearer %s", config.ApiToken)

	data, err := json.Marshal(map[string]interface{}{
		"content": ipv4,
	})

	if err != nil {
		log.Fatalln(err)
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(data))
	req.Header.Add("Authorization", token)

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var response CloudflareVerificationResponse

	json.NewDecoder(resp.Body).Decode(&response)

	if response.Success {
		currentIPv4 = ipv4
		lastUpdated = time.Now().Format(time.RFC1123)
		fmt.Printf("\n\rLast Updated: %s", lastUpdated)
		fmt.Print(UP_LINE)
	} else {
		log.Fatalln(err)
	}
}

func getPublicIPv4() string {
	client := &http.Client{
		CheckRedirect: http.DefaultClient.CheckRedirect,
	}

	url := "https://ifconfig.co/ip"

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	b, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	r := regexp.MustCompile(`((?:[0-9]{1,3}\.){3}[0-9]{1,3})`)

	// ifconfig.co returns an IP address with a newline character at the end, so we're using regex to chop it off
	return r.FindString(string(b))
}

func getTimeFromSeconds(seconds int) string {
	hours := seconds / 3600
	minutes := seconds / 60

	if hours > 0 {
		return fmt.Sprintf("%d hour(s)\u0020", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%d minutes(s)\u0020", minutes)
	} else {
		return fmt.Sprintf("%d second(s)", seconds)
	}
}
