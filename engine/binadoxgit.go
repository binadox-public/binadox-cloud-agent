package engine

import (
	"context"
	"github.com/google/go-github/github"
)

type Asset struct {
	name string
	url  string
}

type Release struct {
	tag         string
	description string
	assets      []Asset
}

func ListReleases() ([]Release, error) {
	client := github.NewClient(nil)
	opt := &github.ListOptions{Page: 1, PerPage: 10}
	var resultReleases []Release

	for {
		releases, rsp, err := client.Repositories.ListReleases(context.Background(), "binadox-public", "cloud-agent", opt)
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
				a := Asset{url: *asset.BrowserDownloadURL, name: *asset.Name}
				assets = append(assets, a)
			}
			out := Release{tag: *r.TagName, description: *r.Body, assets: assets}
			resultReleases = append(resultReleases, out)
		}

		if rsp.NextPage == 0 {
			break
		}
		opt.Page = rsp.NextPage
	}

	return resultReleases, nil
}
