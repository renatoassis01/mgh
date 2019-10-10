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
)

const baseURL = "https://api.github.com/users"

//API base
type API struct {
	Client  *http.Client
	baseURL string
}

type Repository struct {
	FullName string `json:"full_name"`
}

type User struct {
	username string
	repos    []string
	token    string
}

type ErrorMesage struct {
	Message string `json:"message"`
}

func (g *API) GetRepository(user *User) error {
	url := fmt.Sprintf("%s/%s/repos", g.baseURL, user.username)
	log.Printf("Fetching %s", url)
	resp, err := g.Client.Get(url)
	if err != nil {
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

func (g *API) DeleteRepository(user User) {
	url := fmt.Sprintf("%s/%s", g.baseURL, user.username)
	req, err := http.NewRequest("DELETE", url, nil)
	req.Header.Add("Authorization: token", user.token)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := g.Client.Do(req)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(resp.Status)
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(resp.body)
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
	answers := []string{}
	err := survey.Ask(multiQs, &answers, survey.WithPageSize(20))
	if err != nil {
		log.Println(err.Error())
		return
	}
	DeleteRepo := false
	prompt := &survey.Confirm{
		Message: "Do you delete repos",
	}
	survey.AskOne(prompt, &DeleteRepo)
	if DeleteRepo {
		// print the answers
		fmt.Printf("you chose: %s\n", strings.Join(answers, ", "))

		prompt := &survey.Password{
			Message: "Please enter your Github's TOKEN: ",
		}
		survey.AskOne(prompt, &user.token)

	}

}
func main() {
	user := User{}
	github := API{Client: &http.Client{}, baseURL: baseURL}
	flag.StringVar(&user.username, "username", "", "Your username github")
	flag.Parse()

	if len(os.Args) == 1 {
		usage()
	}
	error := github.GetRepository(&user)
	if error != nil {
		log.Fatalln(error)

	}
	Menu(&user)

}

func usage() {
	example := fmt.Sprintf("Example usage of %s -username=jonhdoe", os.Args[0])
	fmt.Println(example)
	os.Exit(1)
}
