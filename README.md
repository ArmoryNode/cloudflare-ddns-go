# cloudflare-ddns
A small application to dynamically update a DNS record in CloudFlare.

# Purpose
I mainly wrote this app as a way to learn how to write Go. It's taking some getting used to, but I feel like I'm getting the hang of things.

# Prerequisites
* Go version 1.19 or above.
* A Cloudflare API key with `DNS:Edit` permission.
* Your zone id for your domain.
* Your DNS record identifier. ([see the "Getting your DNS record identifier" section for instructions on how to get this.](https://github.com/ArmoryNode/cloudflare-ddns#getting-your-dns-record-identifier))

# Build Instructions
* Add information to the `config.json` file.
* Open a terminal at the root of the project and run `go build`.
* Then run `./cloudflare-ddns`.

Once the program verifies your API key and Zone/DNS record identifiers, it will update the DNS record at the specified interval.

# Getting your DNS record identifier
Cloudflare currently does not have an easy way to view the ids for your DNS records. This is the least painful way to get it I've found.

* Go to [this page](https://api.cloudflare.com/#dns-records-for-a-zone-list-dns-records) in Cloudflare's docs.
* Execute the curl command in a terminal, or your preferred REST client. (make sure to strip off the additional query parameters from their example.)
* You can use the same API key that you're using for the project, just be sure it also has the `DNS:Read` permission along with edit permission.
* Look for your DNS record and grab the `"id"` value.

Once you have that you're good to go.
