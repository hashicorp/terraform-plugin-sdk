package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/cmd/importers/github"
	"github.com/hashicorp/terraform-plugin-sdk/cmd/importers/godoc"
	"github.com/hashicorp/terraform-plugin-sdk/cmd/importers/util"
)

var packages = [...]string{
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

type repo struct {
	Stars    int                 `json:"stars"`
	Packages map[string][]string `json:"packages"`
}

func main() {
	client := github.NewClient(context.Background(), util.MustEnv("GITHUB_PERSONAL_TOKEN"))
	r := make(map[string]*repo)

	ignore, err := loadIgnoreSet(client)
	if err != nil {
		log.Fatalf("Error loading set of projects to ignore: %s", err)
	}

	for _, pkg := range packages {
		pkg = "github.com/hashicorp/terraform/" + pkg
		importers, err := godoc.ListImporters(pkg, ignore, true)
		if err != nil {
			log.Fatalf("Error fetching importers of %s: %s", pkg, err)
		}
		for _, imp := range importers {
			// non github repos will have the full package path
			// it will be unclear to us where the project namespace begins
			// and where the package tree begins
			proj := github.RepoRoot(imp)
			if _, ok := r[proj]; !ok {
				var stars int
				if strings.HasPrefix(imp, "github.com") {
					var err error
					owner, repo := github.OwnerRepo(imp)
					stars, err = client.GetStars(owner, repo)
					if err != nil {
						log.Println(err)
					}
				}
				r[proj] = &repo{
					Stars: stars,
					Packages: map[string][]string{
						imp: []string{pkg},
					},
				}
			} else {
				r[proj].Packages[imp] = append(r[proj].Packages[imp], pkg)
			}
		}
	}

	if err := json.NewEncoder(os.Stdout).Encode(r); err != nil {
		log.Fatalf("Error writing report: %s", err)
	}
}

func loadIgnoreSet(client *github.Client) (map[string]bool, error) {
	ignoreForksOf, err := client.ListRepositories("terraform-providers")
	if err != nil {
		return nil, err
	}
	ignoreForksOf = append(ignoreForksOf, "github.com/hashicorp/terraform", "github.com/hashicorp/otto")

	var ignoredForks []string
	for _, upstream := range ignoreForksOf {
		owner, repo := github.OwnerRepo(upstream)
		forks, err := client.ListForks(owner, repo)
		if err != nil {
			return nil, err
		}
		ignoredForks = append(ignoredForks, forks...)
	}
	return util.StringListToSet(append(ignoredForks, ignoreForksOf...)), nil
}
