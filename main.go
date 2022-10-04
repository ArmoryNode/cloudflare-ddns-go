package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ArmoryNode/cloudflare-ddns/cli"
	"github.com/ArmoryNode/cloudflare-ddns/cloudflare"
)

const version = "1.0.2"

// Used as an indicator that the application is running
var spinner = [4]rune{'|', '\\', '-', '/'}

var logo string = ""

func main() {
	// Hide the cursor in the console
	fmt.Print(cli.HIDE_CURSOR)
	fmt.Print(cli.COLOR_GREEN)
	fmt.Print(cli.CLEAR_SCREEN)

	printFrame()

	apiTokenPtr := flag.String("apiToken", "", "Your API token from Cloudflare")
	zoneIdentifierPtr := flag.String("zoneIdentifier", "", "The identifier for your zone (domain)")
	dnsRecordIdentifierPtr := flag.String("dnsRecordIdentifier", "", "The DNS record identifier")
	updateIntervalPtr := flag.Int("updateInterval", 1, "The update interval in minutes")

	flag.Parse()

	if *apiTokenPtr == "" {
		fmt.Printf("%sError: API token required\n", cli.COLOR_YELLOW)
		return
	} else if *zoneIdentifierPtr == "" {
		fmt.Printf("%sError: Zone identifier required\n", cli.COLOR_YELLOW)
		return
	} else if *dnsRecordIdentifierPtr == "" {
		fmt.Printf("%sError: DNS record identifier required\n", cli.COLOR_YELLOW)
		return
	}

	cloudflareClient := cloudflare.CreateClient(*apiTokenPtr, *zoneIdentifierPtr, *dnsRecordIdentifierPtr, *updateIntervalPtr)

	done := make(chan bool)

	fmt.Print(cli.CLEAR_LINE)
	fmt.Println("[✓] Configuration loaded")

	go cli.DisplaySpinner("Verifying Cloudflare API token", done)
	verified := cloudflareClient.VerifyApiToken(done)

	if verified {
		fmt.Print(cli.CLEAR_LINE)
		fmt.Println("[✓] API token verified")
	} else {
		fmt.Print(cli.CLEAR_LINE)
		fmt.Println(cli.COLOR_RED + "[✕] Failed to verify API token")
		return
	}

	go cli.DisplaySpinner("Verifying zone", done)
	verified = cloudflareClient.VerifyZone(done)

	if verified {
		fmt.Print(cli.CLEAR_LINE)
		fmt.Println("[✓] Zone verified")
	} else {
		fmt.Print(cli.CLEAR_LINE)
		fmt.Println(cli.COLOR_RED + "[✕] Failed to verify zone")
		return
	}

	go cli.DisplaySpinner("Verifying DNS record", done)
	verified = cloudflareClient.VerifyDnsRecord(done)

	if verified {
		fmt.Print(cli.CLEAR_LINE)
		fmt.Println("[✓] DNS record verified")
	} else {
		fmt.Print(cli.CLEAR_LINE)
		fmt.Println(cli.COLOR_RED + "[✕] Failed to verify DNS record")
		return
	}

	close(done)

	// Make a channel for signaling an update
	update := make(chan bool)

	str := cloudflareClient.UpdateRecord()
	fmt.Print(cli.CLEAR_SCREEN)
	printFrame(str)

	go countdown(cloudflareClient.UpdateIntervalInMinutes, update)

	for {
		<-update
		str := cloudflareClient.UpdateRecord()
		printFrame(str)
	}
}

func countdown(minutes int, update chan bool) {
	secondsFrom := minutes * 60
	for {
		if secondsFrom >= 0 {
			printFrame(fmt.Sprintf("[%c] Updating in %s", spinner[secondsFrom%4], getTimeFromSeconds(secondsFrom)))
			secondsFrom--
			time.Sleep(time.Second)
		} else {
			secondsFrom = minutes * 60
			update <- true
		}
	}
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

func getLogo() string {
	if logo == "" {
		content, _ := os.ReadFile("logo.txt")
		logo = string(content)
	}

	return fmt.Sprintf(logo+"\tv%s\n\n", version)
}

func printFrame(additionalLines ...string) {
	str := ""

	str += fmt.Sprint(cli.MOVE_CURSOR_HOME)
	str += fmt.Sprint(getLogo())

	for i := 0; i < len(additionalLines); i++ {
		str += fmt.Sprint(additionalLines[i])
	}

	fmt.Print(str)
}
