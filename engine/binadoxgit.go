package engine

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Asset struct {
	Name string
	Url  string
}

type Release struct {
	Tag         string
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
			out := Release{Tag: *r.TagName, Assets: assets}
			resultReleases = append(resultReleases, out)
		}

		if rsp.NextPage == 0 {
			break
		}
		opt.Page = rsp.NextPage
	}

	return resultReleases, nil
}

type TokenSource struct {
    AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
    token := &oauth2.Token{
        AccessToken: t.AccessToken,
    }
    return token, nil
}
