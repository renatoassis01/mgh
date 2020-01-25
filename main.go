package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	. "github.com/logrusorgru/aurora"
)

const baseURL = "https://api.github.com"

//API base
type API struct {
	Client  *http.Client
	baseURL string
}

type Repository struct {
	FullName string `json:"full_name"`
}

type User struct {
	username     string
	repos        []string
	token        string
	DeleteRepos  []string
	StatusDelete bool
}

type ErrorMesage struct {
	Message string `json:"message"`
}

type Result struct {
	Url        string
	StatusCode int
}

func (g *API) GetStatuses(user *User) <-chan Result {
	var urls []string
	ch := make(chan Result)
	for _, repo := range user.DeleteRepos {
		url := fmt.Sprintf("%s/repos/%s", g.baseURL, repo)
		urls = append(urls, url)
	}
	for _, url := range urls {
		go g.DeleteRepository(url, user.token, ch)
	}
	return ch
}

func (g *API) GetRepository(user *User) error {
	url := fmt.Sprintf("%s/user/repos?per_page=200&type=all", g.baseURL)
	log.Printf("Fetching %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	bearer := fmt.Sprintf("Bearer %s", user.token)
	req.Header.Add("Authorization", bearer)
	resp, err := g.Client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var error ErrorMesage
		json.Unmarshal(body, &error)
		return fmt.Errorf("Message from Github's : %v - %s", resp.StatusCode, error.Message)
	}

	var repos []Repository
	json.Unmarshal(body, &repos)
	for _, repo := range repos {
		user.repos = append(user.repos, repo.FullName)

	}

	return nil

}

func (g *API) DeleteRepository(url, token string, ch chan Result) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	bearer := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", bearer)
	resp, err := g.Client.Do(req)
	if err != nil {
		panic(err)
	}
	result := Result{
		Url:        url,
		StatusCode: resp.StatusCode,
	}
	ch <- result

}

func Menu(user *User) {
	var multiQs = []*survey.Question{

		{
			Name: "letter",
			Prompt: &survey.MultiSelect{
				Message: "Select your repository delete :",
				Options: user.repos,
			},
		},
	}

	err := survey.Ask(multiQs, &user.DeleteRepos, survey.WithPageSize(20))
	if err != nil {
		log.Fatalln(err.Error())

	}

	user.StatusDelete = false
	prompt := &survey.Confirm{
		Message: "Do you delete repos",
	}
	survey.AskOne(prompt, &user.StatusDelete)
	if user.StatusDelete {
		fmt.Printf("you chose: %s\n", Red(strings.Join(user.DeleteRepos, ", ")))
	} else {
		os.Exit(0)
	}

}
func main() {
	user := User{}
	flag.StringVar(&user.username, "username", "", "Your username github")
	flag.Parse()

	if len(os.Args) == 1 {
		usage()
	}
	promptp := &survey.Password{
		Message: "Please enter your Github's TOKEN: ",
	}
	survey.AskOne(promptp, &user.token)
	github := API{Client: &http.Client{}, baseURL: baseURL}
	error := github.GetRepository(&user)
	if error != nil {
		log.Fatalln(error)

	}
	Menu(&user)
	ch := github.GetStatuses(&user)
	for i := 0; i < len(user.DeleteRepos); i++ {
		res := <-ch
		fmt.Printf("url: %s - status code: %d\n", res.Url, res.StatusCode)
	}
	fmt.Println("complete.")

}

func usage() {
	example := fmt.Sprintf("Example usage of %s -username=jonhdoe", os.Args[0])
	fmt.Println(example)
	os.Exit(1)
}
