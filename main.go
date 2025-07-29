package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Organisation struct {
	Login   string `json:"login"`
	Url     string `json:"url"`
	RepoUrl string `json:"repos_url"`
}

type Repo struct {
	Name string `json:"name"`
}

const BASE_URL = "https://api.github.com"

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func showLoading(message string, done chan bool) {
	fmt.Printf("%s", message)
	i := 0
	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf(".")
			time.Sleep(500 * time.Millisecond)
			i++
		}
	}
}

func makeRequest(url string, output *[]map[string]any) {
	ghKey := os.Getenv("GH_TOKEN")

	if ghKey == "" {
		fmt.Println("GH_TOKEN is empty, Please set this in ur env variable, Refer this url on how to create gh token")
		os.Exit(0)
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ghKey))
	req.Header.Set("Accept", "application/vnd.github+json")
	if err != nil {
		fmt.Println("error")
	}

	client := &http.Client{}
	response, _ := client.Do(req)

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	var out []map[string]any
	json.Unmarshal(body, &out)
	*output = append(*output, out...)

	headerLink := response.Header.Get("Link")
	link := strings.Split(headerLink, "rel=\"next\"")

	if (!strings.Contains(headerLink, "rel=\"next\"")) || len(link) == 1 && link[0] == "" {
		return
	}

	nextLink := link[0]

	// this is the most dumbest way i could think of, will change soon(may be before i die)
	if strings.Contains(link[0], "prev") {
		nextLink = strings.Split(link[0], "rel=\"prev\", ")[1]
	}
	nextUrl := strings.Trim(strings.TrimSpace(nextLink), "<>;, ")

	if url != nextUrl {
		makeRequest(nextUrl, output)
		return
	}

	return
}

func promptUI(message string, options *[]string) string {
	prompt := &survey.Select{
		Message: message,
		Options: *options,
	}

	var selectedVal string
	err := survey.AskOne(prompt, &selectedVal)

	if err != nil {
		if err == io.EOF {
			fmt.Println("\nInterrupted by user. Exiting.")
			os.Exit(0)
		}
		fmt.Printf("Prompt error: %v\n", err)
		os.Exit(1)
	}

	return selectedVal
}

func executeCommand(gitUrl string, repoName string) (*string, error) {
	var output bytes.Buffer
	dir := os.Getenv("CLONE_DIR")

	if dir == "" {
		dir = fmt.Sprintf("%s/Documents", os.Getenv("HOME"))
		fmt.Printf("CLONE_DIR is not provided fallback to %s \n", dir)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Directory %s does not exist", dir)
	}

	path := fmt.Sprintf("%s/%s", dir, repoName)
	cmd := exec.Command("git", "clone", gitUrl, path)

	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	result := output.String()

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if strings.Contains(result, "Permission") {
				return &result, fmt.Errorf("SSH_KEY_INVALID")
			}
			return &result, fmt.Errorf("REPO_ALREADY_EXIST")
		}
		return &result, fmt.Errorf("git clone failed: %w", err)
	}

	return &result, nil
}

func repoPrompt(repos []string, selectedOrg string, repoDetails []map[string]string, done chan bool) {
	selectedRepo := promptUI("Choose your repo: ", &repos)

	for _, val := range repoDetails {
		if val["name"] == selectedRepo {
			_, err := executeCommand(val["ssh_url"], selectedRepo)

			if err != nil {
				if fmt.Sprintf("%s", err) == "SSH_KEY_INVALID" {
					fmt.Printf("%s\nTrying with HTTP URL", err)
					httpUrl := fmt.Sprintf("https://%s@github.com/%s/%s.git", os.Getenv("GH_TOKEN"), selectedOrg, selectedRepo)
					_, error := executeCommand(httpUrl, selectedRepo)

					if error != nil {
						fmt.Println("Error in Cloning", error)
						return
					}
				} else if fmt.Sprintf("%s", err) == "REPO_ALREADY_EXIST" {
					fmt.Printf("Repo already exist at the given path %s/%s", os.Getenv("CLONE_DIR"), selectedRepo)
					return
				} else {
					fmt.Println(err)
					return
				}
			}
		}
	}

	fmt.Printf("\nCloned Successfully under %s/%s...\n", os.Getenv("CLONE_DIR"), selectedRepo)
	time.Sleep(3 * time.Second)
	clearScreen()
	repoPrompt(repos, selectedOrg, repoDetails, done)
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTSTP)

	go func() {
		sig := <-sigs
		fmt.Printf("\n %s, Exiting...\n", sig)
		os.Exit(0)
	}()

	var output []map[string]any
	var done chan bool

	makeRequest(fmt.Sprintf("%s/user/orgs?per_page=100&page=1", BASE_URL), &output)

	var organisations []string

	for _, val := range output {
		login, ok := val["login"].(string)
		if ok {
			organisations = append(organisations, login)
		} else {
			fmt.Println("login is not a string or missing")
		}
	}
	selectedOrg := promptUI("Choose your org: ", &organisations)

	output = nil
	done = make(chan bool)
	go showLoading("Loading", done)
	makeRequest(fmt.Sprintf("%s/orgs/%s/repos?per_page=100&page=1", BASE_URL, selectedOrg), &output)

	done <- true

	var repoDetails []map[string]string
	var repos []string

	for _, val := range output {
		name, _ := val["name"].(string)
		ssh_url, _ := val["ssh_url"].(string)
		repoDetails = append(repoDetails, map[string]string{
			"name":    name,
			"ssh_url": ssh_url,
		})
		repos = append(repos, name)
	}

	clearScreen()
	repoPrompt(repos, selectedOrg, repoDetails, done)
}
