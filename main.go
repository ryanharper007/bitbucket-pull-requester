package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	bitbucket "github.com/ktrysmt/go-bitbucket"
	"github.com/ryanharper007/go-keyring"
)

// Pullrequest struct type.
type Pullrequest struct {
	Owner               string `json:"owner"`
	Repo_slug           string `json:"repo_slug"`
	Close_source_branch bool   `json:"close_source_branch"`
	Source_branch       string `json:"source_branch"`
	Destination_branch  string `json:"destination_branch"`
	Title               string `json:"title"`
	Message             string `json:"message"`
}

func main() {
	// Setup command line options
	User := flag.String("user", "", "Your Bitbucket username (Required)")
	Source := strings.TrimSuffix(string(runcmd("git rev-parse --abbrev-ref HEAD", true)), "\n")
	GitRemote := string(runcmd("git remote -v | sed -n 2p | awk '{print $2}'", true))
	Dest := flag.String("dest", "", "The target branch you would like to merge this pull request to (Required)")
	Message := flag.String("message", "", "The commit message (Required)")
	Debug := flag.Bool("debug", false, "Debug the command line")
	service := "bitbucket.org"

	// Parse the flags
	flag.Parse()

	// If message is not defined barf
	if *Message == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// If user is not defined barf
	if *User == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// If Dest is not defined barf
	if *Dest == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// If Gitremote is not defined barf
	if GitRemote == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Define the git remote
	s := strings.Split(GitRemote, "/")
	Slug := string(strings.Split(s[len(s)-1], ".")[0])

	// Use your bitbucket secrets to get your password (OSX)
	secret, err := keyring.IntGet(service, *User)
	if err != nil {
		log.Fatal(err)
	}

	// If Debugging is enabled then echo out the commands
	if *Debug {
		println(Source)
		println(User)
		println(secret)
		println(Slug)
		println(Message)
	}

	// Marshal our request into Json, so that the interpolation happens
	s2, _ := json.Marshal(Pullrequest{
		Owner:               "sedex",
		Repo_slug:           Slug,
		Close_source_branch: true,
		Destination_branch:  *Dest,
		Title:               *Message,
		Message:             *Message,
		Source_branch:       Source,
	})

	// Pretty Print our json
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, s2, "", "\t")
	if error != nil {
		log.Println("JSON parse error: ", error)
		return
	}

	// Unmarshal our json into a struct that the pullrequester will consume
	pullrequestMap := &bitbucket.PullRequestsOptions{}
	json.Unmarshal([]byte(prettyJSON.Bytes()), &pullrequestMap)

	// Instantiate auth
	c := bitbucket.NewBasicAuth(*User, secret)

	// Create the pull request
	res, err := c.Repositories.PullRequests.Create(pullrequestMap)
	if err != nil {
		panic(err)

	}
	// Print out the result
	fmt.Println(res)

	//resMap := make(map[string]interface{})
	blah, _ := json.Marshal(res)

	var prettyJSON2 bytes.Buffer
	error2 := json.Indent(&prettyJSON2, blah, "", "\t")
	if error2 != nil {
		log.Println("JSON parse error: ", error)
		return
	}
	fmt.Printf("%s", &prettyJSON2)
	//fmt.Println(blah)
}

func runcmd(cmd string, shell bool) []byte {
	if shell {
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			log.Fatal(err)
			panic("some error found")
		}
		return out
	}
	out, err := exec.Command(cmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	return out
}
