package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v26/github"
	"golang.org/x/oauth2"
)

var packages []string = []string{
	"helper/acctest",
	"helper/customdiff",
	"helper/encryption",
	"helper/hashcode",
	"helper/logging",
	"helper/mutexkv",
	"helper/pathorcontents",
	"helper/resource",
	"helper/schema",
	"helper/structure",
	"helper/validation",
	"httpclient",
	"plugin",
	"terraform",
}

type Importer struct {
	Path string `json:"path"`
}

type Importers struct {
	Results []Importer `json:"results"`
}

func mustEnv(env string) string {
	v := os.Getenv(env)
	if v == "" {
		log.Fatalf("Env Var %q must be set", env)
	}
	return v
}

func main() {
	forks := getForks()
	imps := make(map[string]bool)
	for _, pkg := range packages {
		for _, p := range getImporters(pkg, forks) {
			imps[p] = true
		}
	}
	for repo, _ := range imps {
		fmt.Println(repo)
	}
}

func getImporters(pkg string, forks map[string]bool) []string {
	url := "https://api.godoc.org/importers/github.com/hashicorp/terraform/" + pkg
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error getting %q: %s", url, err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %s", err)
	}

	var imps Importers
	if err := json.Unmarshal(body, &imps); err != nil {
		log.Fatalf("Error parsing json: %s", err)
	}

	var res []string
	for _, imp := range imps.Results {
		parts := strings.Split(imp.Path, "/")
		if len(parts) < 3 {
			fmt.Printf("Strange import of package: %s\n", imp.Path)
			continue
		}
		repo := strings.Join(parts[:3], "/")
		if !forks[repo] && !strings.Contains(repo, "gopkg.in") {
			res = append(res, repo)
		}
	}
	return res
}

// get all forks of hashicorp/terraform and terraform-providrs/*
func getForks() map[string]bool {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: mustEnv("GITHUB_PERSONAL_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	forks := make(map[string]bool)

	forks["github.com/hashicorp/terraform"] = true
	// get forks of hashicorp/terraform
	opt := &github.RepositoryListForksOptions{ListOptions: github.ListOptions{PerPage: 200}}
	for {
		repos, resp, err := client.Repositories.ListForks(ctx, "hashicorp", "terraform", opt)
		if err != nil {
			log.Fatalf("Could not retrieve forks: %s", err)
		}
		for _, repo := range repos {
			forks["github.com/"+repo.GetFullName()] = true
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// get terraform-providers/*
	opt2 := &github.RepositoryListByOrgOptions{Type: "public", ListOptions: github.ListOptions{PerPage: 200}}
listproviders:
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, "terraform-providers", opt2)
		if err != nil {
			log.Fatalf("Could not retrieve providers: %s", err)
		}
		for _, repo := range repos {
			parts := strings.Split(repo.GetFullName(), "/")
			owner := parts[0]
			name := parts[1]
			forks["github.com/"+repo.GetFullName()] = true
			// get forks
			opt3 := &github.RepositoryListForksOptions{ListOptions: github.ListOptions{PerPage: 200}}
		listforks:
			for {
				repos, resp, err := client.Repositories.ListForks(ctx, owner, name, opt3)
				if err != nil {
					log.Fatalf("Could not retrieve forks: %s", err)
				}
				for _, repo := range repos {
					forks["github.com/"+repo.GetFullName()] = true
				}
				if resp.NextPage == 0 {
					break listforks
				}
				opt3.Page = resp.NextPage
			}
		}

		if resp.NextPage == 0 {
			break listproviders
		}
		opt2.Page = resp.NextPage
	}

	return forks
}
