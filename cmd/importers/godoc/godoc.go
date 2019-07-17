package godoc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/cmd/importers/github"
)

type importer struct {
	Path string `json:"path"`
}

type importers struct {
	Results []importer `json:"results"`
}

func ListImporters(pkg string, ignore map[string]bool, ignoreGopkg bool) ([]string, error) {
	var res []string
	url := "https://api.godoc.org/importers/" + pkg
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error getting %q: %s", url, err)
	}
	defer resp.Body.Close()

	var imps importers
	if err := json.NewDecoder(resp.Body).Decode(&imps); err != nil {
		return res, fmt.Errorf("Error decoding response: %s", err)
	}

	for _, imp := range imps.Results {
		if ignore[github.RepoRoot(imp.Path)] {
			continue
		}
		if ignoreGopkg && strings.HasPrefix(imp.Path, "gopkg.in") {
			continue
		}
		res = append(res, imp.Path)
	}
	return res, nil
}
