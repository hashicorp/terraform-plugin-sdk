package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v26/github"
	"golang.org/x/oauth2"
)

const (
	domain = "github.com/"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient(ctx context.Context, token string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
	}
}

func (c *Client) GetStars(owner, repo string) (int, error) {
	r, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		return 0, fmt.Errorf("Error getting stars for %s/%s: %s", owner, repo, err)
	}
	return r.GetStargazersCount(), nil
}

func (c *Client) ListRepositories(owner string) ([]string, error) {
	var repos []string
	opt := &github.RepositoryListByOrgOptions{Type: "public", ListOptions: github.ListOptions{PerPage: 200}}
	for {
		r, resp, err := c.client.Repositories.ListByOrg(c.ctx, owner, opt)
		if err != nil {
			return repos, fmt.Errorf("Could not retrieve repos of %s/: %s", owner, err)
		}
		for _, repo := range r {
			repos = append(repos, domain+repo.GetFullName())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return repos, nil
}

func (c *Client) ListForks(owner, repo string) ([]string, error) {
	var repos []string
	opt := &github.RepositoryListForksOptions{ListOptions: github.ListOptions{PerPage: 200}}
	for {
		r, resp, err := c.client.Repositories.ListForks(c.ctx, owner, repo, opt)
		if err != nil {
			return repos, fmt.Errorf("Could not retrieve forks of %s/%s: %s", owner, repo, err)
		}
		for _, repo := range r {
			repos = append(repos, domain+repo.GetFullName())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return repos, nil
}

func RepoRoot(pkg string) string {
	parts := strings.Split(pkg, "/")
	if len(parts) < 3 {
		return pkg
	} else {
		return strings.Join(parts[:3], "/")
	}
}

func OwnerRepo(pkg string) (string, string) {
	parts := strings.Split(pkg, "/")
	return parts[1], parts[2]
}
