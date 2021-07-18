package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Asset struct {
	Name string
	Url  string
}

type Release struct {
	Tag         string
	Description string
	Assets      []Asset
}

func ListReleases() ([]Release, error) {
	client := github.NewClient(nil)
	opt := &github.ListOptions{Page: 1, PerPage: 10}
	var resultReleases []Release

	for {
		releases, rsp, err := client.Repositories.ListReleases(context.Background(), "binadox-public", "binadox-cloud-agent", opt)
		if err != nil {
			return nil, err
		}

		for i, _ := range releases {
			r := releases[i]
			if r.Prerelease != nil && *r.Prerelease {
				continue
			}
			if r.TagName == nil {
				continue
			}
			if r.Body == nil {
				continue
			}
			var assets []Asset
			for j, _ := range r.Assets {
				asset := r.Assets[j]
				if asset.URL == nil {
					continue
				}
				if asset.Name == nil {
					continue
				}
				a := Asset{Url: *asset.BrowserDownloadURL, Name: *asset.Name}
				assets = append(assets, a)
			}
			out := Release{Tag: *r.TagName, Description: *r.Body, Assets: assets}
			resultReleases = append(resultReleases, out)
		}

		if rsp.NextPage == 0 {
			break
		}
		opt.Page = rsp.NextPage
	}

	return resultReleases, nil
}
// ----------------------------------------------------------------------------------
type TokenSource struct {
    AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
    token := &oauth2.Token{
        AccessToken: t.AccessToken,
    }
    return token, nil
}
func TestUpload(personalAccessToken string)  {
	tokenSource := &TokenSource{
        AccessToken: personalAccessToken,
    }
    oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
    client := github.NewClient(oauthClient)
    client.Repositories.CreateRelease()
    user, _, err := client.Users.Get(context.TODO(), "")
    if err != nil {
        fmt.Printf("client.Users.Get() faled with '%s'\n", err)
        return
    }
    d, err := json.MarshalIndent(user, "", "  ")
    if err != nil {
        fmt.Printf("json.MarshlIndent() failed with %s\n", err)
        return
    }
    fmt.Printf("User:\n%s\n", string(d))
}