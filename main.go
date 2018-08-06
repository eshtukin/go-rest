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
	"strconv"
	"time"

	"github.com/mrjones/oauth"
)

const (
	baseUrlGitHub     = "https://api.github.com/repos/"
	userRepoGutHub    = "eshtukin/go-rest/"
	pullsRelUrlGitHub = "pulls?direction=asc"

	baseUrlTwitter    = "https://api.twitter.com/1.1/"
	postRelUrlTwitter = "statuses/update.json?status="
	tweetLimit        = 280

	baseLineFileName = "baseline.txt"
	timeStampLayout  = "2006-01-02 15:04:05"
	initialBaseLine  = "2006-01-02 15:04:05"
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

var gitHubClient *http.Client
var twitterClient *http.Client

var baseLineFile *os.File

var (
	consumerKeyTwitterAPI       = getEnv("TWITTER_CONSUMER_KEY")
	consumerSecretKeyTwitterAPI = getEnv("TWITTER_CONSUMER_SECRET")
	accessTokenTwitter          = getEnv("TWITTER_ACCESS_TOKEN")
	accessTokenSecretTwitter    = getEnv("TWITTER_ACCESS_TOKEN_SECRET")
)

func getEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Fatal("missing environment valirable " + name)
	}
	return val
}

func ProcessPullRequests(w http.ResponseWriter, r *http.Request) {

	// Get full details about given repository Pull Requests
	getPullsFromGitHub()

	// Figure out timestamp used in previous run
	openBaseLineFile()
	defer baseLineFile.Close()
	prevTime := bringPrevBaseLine()
	currTime := time.Now()

	// Loop through all open pull requests
	for _, item := range pulls {
		fmt.Printf("%+v \n", item)
		// filtering out old pull requests
		if item.CreatedAt.After(prevTime) {
			tweet := constructTweet(&item)
			postTweet(tweet)
		}
	}
	// Keep in the file new baseline for future runs
	saveNewBaseLine(&currTime)
}

func getPullsFromGitHub() {

	// Construct entire URL string
	var buff bytes.Buffer
	buff.WriteString(baseUrlGitHub)
	buff.WriteString(userRepoGutHub)
	buff.WriteString(pullsRelUrlGitHub)
	urlString := buff.String()

	// Create and send GET request
	req, err := http.NewRequest("GET", urlString, nil)
	req.Header.Add("content-type", `application/json`)
	// TODO:  add Authorization header - to get 5000 GitHub requests per hour
	response, err := gitHubClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Fill in PullRequest array with only required data
	buf, _ := ioutil.ReadAll(response.Body)
	// TODO:  add pagination handling
	err = json.Unmarshal(buf, &pulls)
	if err != nil {
		log.Fatal(err)
	}
}

func openBaseLineFile() {

	// Open file for read/write or create if not exists
	f, err := os.OpenFile(baseLineFileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	baseLineFile = f
}

func bringPrevBaseLine() time.Time {

	// If exists - take timestamp of previous run from the file
	var timeStr string
	buf, _ := ioutil.ReadAll(baseLineFile)
	if len(buf) > 0 {
		// Previous timestamp from the file
		timeStr = string(buf)
	} else {
		// First run - nothing yet in the file
		timeStr = initialBaseLine
	}

	t, err := time.Parse(timeStampLayout, timeStr)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func saveNewBaseLine(time *time.Time) {

	// Override file content from the beginning
	_, err := baseLineFile.WriteAt([]byte(time.Format(timeStampLayout)), 0)
	if err != nil {
		log.Fatal(err)
	}
}

func constructTweet(item *PullRequest) string {

	// Summary PR string comprising PR number, CreateAt, and Title values
	tweet := "PR#: " + strconv.Itoa(item.Number) + "; CreatedAt: " + item.CreatedAt.Format(timeStampLayout) + "; Title: " + item.Title

	// Retain <= 280 characters allowed on Twitter
	result := []rune(tweet)
	tweetLength := len(result)
	if tweetLength > tweetLimit {
		tweetLength = tweetLimit
	}
	return string(result[:tweetLength])
}

func postTweet(tweet string) {

	// Construct entire URL string
	var buff bytes.Buffer
	buff.WriteString(baseUrlTwitter)
	buff.WriteString(postRelUrlTwitter)
	// Spaces in tweet message should be encoded with '+' or '%20'
	buff.WriteString(url.QueryEscape(tweet))
	urlString := buff.String()

	// Create and send POST request
	req, err := http.NewRequest("POST", urlString, nil)
	req.Header.Add("content-type", `application/json`)
	response, err := twitterClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
}

func setGitHubClient() {
	gitHubClient = &http.Client{}
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

	setGitHubClient()
	setTwitterClient()

	http.HandleFunc("/pulls", ProcessPullRequests)

	log.Fatal(http.ListenAndServe(":8002", nil))
}
