package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pkg/errors"
)

const org = "kronologic"

var (
	linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`) // https://github.com/cli/cli/blob/trunk/pkg/cmd/api/pagination.go
	owners = []string{"drmaples", "chris-kronologic", "v0vanec"}

	client = req.C().
		SetBaseURL("https://api.github.com").
		SetCommonHeader("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN"))).
		SetTimeout(time.Second * 30).
		SetCommonDumpOptions(&req.DumpOptions{
			// Output:         os.Stdout,
			// RequestHeader:  true,
			// ResponseBody:   true,
			// RequestBody:    false,
			// ResponseHeader: false,
			// Async:          false,
		}).EnableDumpAll().
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			if !resp.IsSuccessState() && resp.StatusCode != http.StatusNotFound {
				slog.Error("request failed", slog.String("url", resp.Request.RawURL), slog.Int("status.code", resp.StatusCode))
				return errors.New(resp.String())
			}
			return nil
		})
)

type githubRepoResponse struct {
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
}

type githubRepoTeamResponse struct {
	Name       string `json:"name"`
	Permission string `json:"permission"`
}

type githubRepoCollabResponse struct {
	Login    string `json:"login"`
	Type     string `json:"type"`
	RoleName string `json:"role_name"`
}

type githubRepoBranchProtectionResponse struct {
	RequiredStatusChecks struct {
		Contexts []string `json:"contexts"`
		Checks   []struct {
			Context string `json:"context"`
		} `json:"checks"`
	} `json:"required_status_checks"`
	RequiredPullRequestReviews struct {
		DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
		RequireCodeOwnerReviews      bool `json:"require_code_owner_reviews"`
		RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
	} `json:"required_pull_request_reviews"`
	Restrictions struct {
		URL   string `json:"url"`
		Teams []struct {
			Name string `json:"name"`
		} `json:"teams"`
	} `json:"restrictions"`
}

func findNextPage(h *http.Header) (string, bool) {
	for _, m := range linkRE.FindAllStringSubmatch(h.Get("Link"), -1) {
		if len(m) > 2 && m[2] == "next" {
			return m[1], true
		}
	}
	return "", false
}

func getRepos(ctx context.Context) ([]githubRepoResponse, error) {
	var allData []githubRepoResponse
	url := fmt.Sprintf("/orgs/%s/repos?per_page=100&sort=full_name", org)
	for {
		var rawData []githubRepoResponse
		res, err := client.R().SetContext(ctx).
			SetSuccessResult(&rawData).
			Get(url)
		if err != nil {
			return nil, err
		}

		allData = append(allData, rawData...)

		nextPageURL, hasNextPage := findNextPage(&res.Header)
		if !hasNextPage {
			break
		}
		url = nextPageURL
	}
	return allData, nil
}

func getRepoTeamData(ctx context.Context, repo string) ([]githubRepoTeamResponse, error) {
	url := fmt.Sprintf("/repos/%s/%s/teams", org, repo)
	var rawData []githubRepoTeamResponse
	_, err := client.R().SetContext(ctx).
		SetSuccessResult(&rawData).
		Get(url)
	if err != nil {
		return nil, err
	}
	return rawData, nil
}

func getRepoBranchProtectionData(ctx context.Context, repo string, defaultBranch string) (githubRepoBranchProtectionResponse, error) {
	url := fmt.Sprintf("/repos/%s/%s/branches/%s/protection", org, repo, defaultBranch)
	var rawData githubRepoBranchProtectionResponse
	_, err := client.R().SetContext(ctx).
		SetSuccessResult(&rawData).
		Get(url)
	if err != nil {
		return githubRepoBranchProtectionResponse{}, err
	}
	return rawData, nil
}

// getExtraAdmins returns repo admins that are not org admins
func getExtraAdmins(ctx context.Context, repo string) ([]string, error) {
	url := fmt.Sprintf("/repos/%s/%s/collaborators", org, repo)
	var rawData []githubRepoCollabResponse

	_, err := client.R().SetContext(ctx).
		SetSuccessResult(&rawData).
		Get(url)
	if err != nil {
		return nil, err
	}

	extraAdmins := []string{}
	for _, rd := range rawData {
		if rd.RoleName != "admin" {
			continue
		}
		if slices.Contains(owners, rd.Login) {
			continue
		}
		extraAdmins = append(extraAdmins, rd.Login)
	}
	return extraAdmins, nil
}

func doWork(ctx context.Context) error {
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) == 0 {
		return errors.Errorf("missing GITHUB_TOKEN env var, setup a github PAT with 'repo' permissions, set one up here: https://github.com/settings/tokens")
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"repo", "default branch", "teams", "extra admins", "req checks", "num approvals", "dismiss stale", "code owner review", "restricted", "restricted pushers"})

	repos, err := getRepos(ctx)
	if err != nil {
		return err
	}

	fmt.Print("fetching repo info")
	for _, repo := range repos {
		fmt.Print(".")
		teamData, err := getRepoTeamData(ctx, repo.Name)
		if err != nil {
			return err
		}

		var teams []string
		for _, td := range teamData {
			teams = append(teams, fmt.Sprintf("[%s] %s", td.Permission, td.Name))
		}

		protectionData, err := getRepoBranchProtectionData(ctx, repo.Name, repo.DefaultBranch)
		if err != nil {
			return err
		}
		var restrictedPushers []string
		for _, t := range protectionData.Restrictions.Teams {
			restrictedPushers = append(restrictedPushers, t.Name)
		}

		extraAdmins, err := getExtraAdmins(ctx, repo.Name)
		if err != nil {
			return err
		}

		t.AppendRow([]interface{}{
			repo.Name,
			repo.DefaultBranch,
			strings.Join(teams, "\n"),
			strings.Join(extraAdmins, "\n"),
			protectionData.RequiredStatusChecks.Checks,
			protectionData.RequiredPullRequestReviews.RequiredApprovingReviewCount,
			protectionData.RequiredPullRequestReviews.DismissStaleReviews,
			protectionData.RequiredPullRequestReviews.RequireCodeOwnerReviews,
			len(protectionData.Restrictions.URL) > 0,
			strings.Join(restrictedPushers, "\n"),
		})
		t.AppendSeparator()
	}

	fmt.Println()
	fmt.Println(t.Render())
	return nil
}

func main() {
	if err := doWork(context.Background()); err != nil {
		panic(err)
	}
}
