package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

var (
	ch = make(chan Result)

	ResultFake     = []string{"foo/repo01", "foo/repo02", "foo/repo03"}
	ResposeGitFake = []interface{}{
		map[string]interface{}{
			"full_name": "foo/repo01",
		},
		map[string]interface{}{
			"full_name": "foo/repo02",
		},
		map[string]interface{}{
			"full_name": "foo/repo03",
		},
	}
	ResponseGitStatus = Result{
		Url:        fmt.Sprintf("https://fake.git.server/repos/%s", ResultFake[0]),
		StatusCode: 202,
	}

	user = User{username: "foo", token: "tkonsdsfwekennnnn"}
	api  = API{Client: &http.Client{}, baseURL: "https://fake.git.server"}
)

func TestGetRepository(t *testing.T) {
	GetUrl := "/users/repos?per_page=200&type=all"
	defer gock.Off()
	gock.New("https://fake.git.server").
		Get(GetUrl).
		Reply(200).
		JSON(ResposeGitFake)

	api.GetRepository(&user)
	assert.Equal(t, ResultFake, user.repos)

	// Verify that we don't have pending mocks
	assert.Exactly(t, gock.IsDone(), true)
}
func TestDeleteRepository(t *testing.T) {

	DeleteUrl := fmt.Sprintf("/repos/%s", ResultFake[0])
	defer gock.Off()
	gock.New("https://fake.git.server").
		Delete(DeleteUrl).
		Reply(204).
		JSON(ResposeGitFake)

	ch := api.GetStatuses(&user)
	res := <-ch
	assert.Equal(t, ResponseGitStatus, res)

	// Verify that we don't have pending mocks
	assert.Exactly(t, gock.IsDone(), true)
}
