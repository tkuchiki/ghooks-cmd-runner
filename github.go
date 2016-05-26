package main

import (
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"strings"
)

type githubClient struct {
	client *github.Client
	owner  string
	repo   string
	ref    string
}

func NewClient(owner, repo, ref, token string) githubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return githubClient{
		owner:  owner,
		repo:   repo,
		ref:    ref,
		client: github.NewClient(tc),
	}
}

func createStatus(client *github.Client, owner, repo, ref string, status *github.RepoStatus) error {
	_, _, err := client.Repositories.CreateStatus(owner, repo, ref, status)

	return err
}

func targetURL(g githubClient) string {
	return fmt.Sprintf("https://github.com/%s/%s/commit/%s", g.owner, g.repo, g.ref)
}

func (g githubClient) pendingStatus() error {
	status := NewRepoStatus("pending", targetURL(g), "The build is pending")

	return createStatus(g.client, g.owner, g.repo, g.ref, status)
}

func (g githubClient) successStatus(target string) error {
	if target == "" {
		target = targetURL(g)
	}

	log.Info(target)

	status := NewRepoStatus("success", target, "The build succeeded!")

	return createStatus(g.client, g.owner, g.repo, g.ref, status)
}

func (g githubClient) failureStatus(target string) error {
	if target == "" {
		target = targetURL(g)
	}

	log.Info(target)

	status := NewRepoStatus("failure", target, "The build failed!")

	return createStatus(g.client, g.owner, g.repo, g.ref, status)
}

func NewRepoStatus(state, target, description string) *github.RepoStatus {
	return &github.RepoStatus{
		State:       &state,
		TargetURL:   &target,
		Description: &description,
	}
}

func includeActions(action string, includes []string) bool {
	if len(includes) == 0 {
		return true
	}

	for _, i := range includes {
		if action == i {
			return true
		}
	}

	return false
}

func excludeActions(action string, excludes []string) bool {
	if len(excludes) == 0 {
		return false
	}

	for _, e := range excludes {
		if action == e {
			return true
		}
	}

	return false
}

func parseBranch(payload interface{}) string {
	j := payload.(map[string]interface{})
	if _, ok := j["ref"]; !ok {
		return ""
	}

	branches := strings.SplitN(j["ref"].(string), "/", 3)

	if len(branches) != 3 {
		return ""
	}

	return branches[2]
}

func parseAction(payload interface{}) string {
	j := payload.(map[string]interface{})
	if _, ok := j["action"]; !ok {
		return ""
	}

	return j["action"].(string)
}

func parsePullRequestStatus(payload interface{}) (string, string, string) {
	j := payload.(map[string]interface{})
	if _, ok := j["pull_request"]; !ok {
		return "", "", ""
	}

	s := j["pull_request"].(map[string]interface{})["_links"].(map[string]interface{})["statuses"].(map[string]interface{})["href"].(string)
	statuses := strings.Split(s, "/")

	return statuses[4], statuses[5], statuses[7]
}
