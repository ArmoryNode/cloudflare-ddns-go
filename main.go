package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/armorynode/cloudflare-ddns/cli"
	"github.com/armorynode/cloudflare-ddns/cloudflare"
)

const version = "1.0.0"

// Used as an indicator that the application is running
var spinner = [4]rune{'|', '\\', '-', '/'}

func main() {
	// Hide the cursor in the console
	fmt.Print(cli.HIDE_CURSOR)

	printLogo()

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

	fmt.Print(cli.COLOR_GREEN)

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

	fmt.Println()

	// Make a channel for signaling an update
	update := make(chan bool)

	cloudflareClient.UpdateRecord()

	go countdown(cloudflareClient.UpdateIntervalInMinutes, update)

	for {
		<-update
		cloudflareClient.UpdateRecord()
	}
}

func countdown(minutes int, update chan bool) {
	secondsFrom := minutes * 60
	for {
		if secondsFrom >= 0 {
			fmt.Print(cli.CLEAR_LINE)
			fmt.Printf("\r[%c] Updating in %s", spinner[secondsFrom%4], getTimeFromSeconds(secondsFrom))
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

func printLogo() {
	logo := `
   ________                ________              
  / ____/ /___  __  ______/ / __/ /___  ________ 
 / /   / / __ \/ / / / __  / /_/ / __ \/ ___/ _ \
/ /___/ / /_/ / /_/ / /_/ / __/ / /_/ / /  /  __/
\____/_/\____/\__,_/\__,_/_/ /_/\__,_/_/   \___/ 
    ____  ____  _   _______                      
   / __ \/ __ \/ | / / ___/                      
  / / / / / / /  |/ /\__ \                       
 / /_/ / /_/ / /|  /___/ /                       
/_____/_____/_/ |_//____/ v%s                        
					
	`

	fmt.Printf(logo+"\n", version)
}
