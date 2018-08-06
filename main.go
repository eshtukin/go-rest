package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/mrjones/oauth"
)

const (
	baseUrlTwitter    = "https://api.twitter.com/1.1/"
	postRelUrlTwitter = "statuses/update.json?status="
)

var (
	consumerKeyTwitterAPI       = getEnv("TWITTER_CONSUMER_KEY")
	consumerSecretKeyTwitterAPI = getEnv("TWITTER_CONSUMER_SECRET")
	accessTokenTwitter          = getEnv("TWITTER_ACCESS_TOKEN")
	accessTokenSecretTwitter    = getEnv("TWITTER_ACCESS_TOKEN_SECRET")
)

func getEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		panic("missing environment valirable " + name)
	}
	return val
}

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

	response, _ := http.Get("https://api.github.com/repos/eshtukin/go-rest/pulls")
	buf, _ := ioutil.ReadAll(response.Body)

	err := json.Unmarshal(buf, &pulls)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n %s  %v \n", "How many PRs?     ", len(pulls))
//	fmt.Printf("%v \n", len(pulls))

	// temp:  replace with read from file
	prevTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	for _, item := range pulls {
		fmt.Printf("%+v \n", item)

		if item.CreatedAt.After(prevTime) {
			// Construct new tweet
			postTweet(" PR id: " + strconv.FormatInt(item.ID) + "   Title: " + item.Title)
		}
	}

	fmt.Printf("%s \n", "End")

	// write current time to a file for future calls
	// currTime := time.Now()
}

func postTweet(tweet string) {
	var buff bytes.Buffer
	buff.WriteString(baseUrlTwitter)
	buff.WriteString(postRelUrlTwitter)
	buff.WriteString(url.QueryEscape(tweet))
	urlString := buff.String()
	fmt.Printf("%s \n", urlString)

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
