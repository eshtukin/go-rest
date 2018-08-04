package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mrjones/oauth"
)

const (
	consumerKeyTwitterAPI       = "z78iW4hDPxogvzFPrVE2yK9f4"
	consumerSecretKeyTwitterAPI = "xPAVry5TWGcQbkfqjTSyNzBRqg6nalFUGNVkAdcwkWfZNDEdbW"
	accessTokenTwitter          = "1025108478107377664-OKOvtFSz4cbaHwgngFbYon582PBt8m"
	accessTokenSecretTwitter    = "cILWcTmd3Fxh7oqVzvtXP1O7vUknnXBrFED1XCOFbC778"

	baseUrlTwitter    = "https://api.twitter.com/1.1/"
	postRelUrlTwitter = "statuses/update.json?status="
)

// PullRequest represents a GitHub pull request on a repository.
type PullRequest struct {
	ID        int64      `json:"id,omitempty"`
	Number    int        `json:"number,omitempty"`
	State     string     `json:"state,omitempty"`
	Title     string     `json:"title,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

var pulls []PullRequest
var twitterClient *http.Client

func GetPullRequests(w http.ResponseWriter, r *http.Request) {

	response, _ := http.Get("https://api.github.com/repos/dotnet/cli/pulls")
	buf, _ := ioutil.ReadAll(response.Body)

	err := json.Unmarshal(buf, &pulls)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n %s", "How many PRs?     ")
	fmt.Printf("%v \n", len(pulls))

	// temp:  replace with read from file
	prevTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	for _, item := range pulls {
		fmt.Printf("%+v \n", item)

		if item.CreatedAt.After(prevTime) {

			// Construct new tweet
			postTweet("Second test of this Twitter account")
			break
		}
	}

	fmt.Printf("%s \n", "End")

	// write current time to a file for future calls
	// currTime := time.Now()
}

func postTweet(tweet string) {

	// –url ‘https://api.twitter.com/1.1/statuses/update.json?status=Test%20tweet.%20Setting%20up%20account’
	// –header ‘authorization: OAuth oauth_consumer_key=”YOUR_CONSUMER_KEY”,
	// 							  oauth_nonce=”AUTO_GENERATED_NONCE”,
	// 							  oauth_signature=”AUTO_GENERATED_SIGNATURE”,
	// 							  oauth_signature_method=”HMAC-SHA1”,
	// 							  oauth_timestamp=”AUTO_GENERATED_TIMESTAMP”,
	// 							  oauth_token=”USERS_ACCESS_TOKEN”,
	// 							  oauth_version=”1.0”’
	// –header ‘content-type: application/json’`

	// req, err := http.NewRequest("POST", "http://example.com", nil)
	// // ...
	// req.Header.Add("If-None-Match", `W/"wyzzy"`)
	// resp, err := client.Do(req)

	var buff bytes.Buffer
	buff.WriteString(baseUrlTwitter)
	buff.WriteString(postRelUrlTwitter)
	buff.WriteString(url.QueryEscape(tweet))
	urlString := buff.String()

	fmt.Printf("%s \n", urlString)

	// u, err := url.Parse(urlString)

	// //	u, err := url.Parse("http://example.com/path with spaces")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(u.EscapedPath())

	//	url = url.EscapedPath()

	req, err := http.NewRequest("POST", urlString, nil)
	req.Header.Add("content-type", `application/json`)
	response, err := twitterClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
}

func setTwitterClient() {

	c := oauth.NewConsumer(
		consumerKeyTwitterAPI,
		consumerSecretKeyTwitterAPI,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		})
	c.Debug(true)

	t := oauth.AccessToken{
		Token:  accessTokenTwitter,
		Secret: accessTokenSecretTwitter,
	}

	client, err := c.MakeHttpClient(&t)
	if err != nil {
		log.Fatal(err)
	}

	twitterClient = client
}

func main() {

	setTwitterClient()

	http.HandleFunc("/pulls", GetPullRequests)

	log.Fatal(http.ListenAndServe(":8002", nil))

}
