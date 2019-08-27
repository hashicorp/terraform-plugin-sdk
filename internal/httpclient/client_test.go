package httpclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/internal/version"
)

func TestUserAgentString_env(t *testing.T) {
	expectedBase := fmt.Sprintf(userAgentFormat, version.Version)
	if oldenv, isSet := os.LookupEnv(uaEnvVar); isSet {
		defer os.Setenv(uaEnvVar, oldenv)
	} else {
		defer os.Unsetenv(uaEnvVar)
	}

	for i, c := range []struct {
		expected   string
		additional string
	}{
		{expectedBase, ""},
		{expectedBase, " "},
		{expectedBase, " \n"},

		{fmt.Sprintf("%s test/1", expectedBase), "test/1"},
		{fmt.Sprintf("%s test/2", expectedBase), "test/2 "},
		{fmt.Sprintf("%s test/3", expectedBase), " test/3 "},
		{fmt.Sprintf("%s test/4", expectedBase), "test/4 \n"},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if c.additional == "" {
				os.Unsetenv(uaEnvVar)
			} else {
				os.Setenv(uaEnvVar, c.additional)
			}

			actual := UserAgentString()

			if c.expected != actual {
				t.Fatalf("Expected User-Agent '%s' does not match '%s'", c.expected, actual)
			}
		})
	}
}

func TestNew_userAgent(t *testing.T) {
	var actualUserAgent string
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		actualUserAgent = req.UserAgent()
	}))
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	for i, c := range []struct {
		expected string
		request  func(c *http.Client) error
	}{
		{
			fmt.Sprintf("Terraform/%s", version.Version),
			func(c *http.Client) error {
				_, err := c.Get(ts.URL)
				return err
			},
		},
		{
			"foo/1",
			func(c *http.Client) error {
				req := &http.Request{
					Method: "GET",
					URL:    tsURL,
					Header: http.Header{
						"User-Agent": []string{"foo/1"},
					},
				}
				_, err := c.Do(req)
				return err
			},
		},
		{
			"",
			func(c *http.Client) error {
				req := &http.Request{
					Method: "GET",
					URL:    tsURL,
					Header: http.Header{
						"User-Agent": []string{""},
					},
				}
				_, err := c.Do(req)
				return err
			},
		},
	} {
		t.Run(fmt.Sprintf("%d %s", i, c.expected), func(t *testing.T) {
			actualUserAgent = ""
			cli := New()
			err := c.request(cli)
			if err != nil {
				t.Fatal(err)
			}
			if actualUserAgent != c.expected {
				t.Fatalf("actual User-Agent '%s' is not '%s'", actualUserAgent, c.expected)
			}
		})
	}
}
