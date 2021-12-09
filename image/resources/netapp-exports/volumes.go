package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type API struct {
	Client   *http.Client
	BaseURL  string
	User     string
	Password string
}

type VolumePathList struct {
	Paths    []string
	NextPage string
}

func resolveURL(base string, ref string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	refURL, err := url.Parse(ref)
	if err != nil {
		return "", err
	}

	return baseURL.ResolveReference(refURL).String(), nil
}

func (api *API) FetchPage() (result VolumePathList, err error) {
	url, err := resolveURL(api.BaseURL, "storage/volumes?nas.path=!null&fields=nas.path&max_records=100")
	if err != nil {
		return
	}
	result, err = api.FetchNextPage(url)
	return
}

func (api *API) FetchNextPage(url string) (result VolumePathList, err error) {
	if url == "" {
		err = fmt.Errorf("url not specified")
		return
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(api.User, api.Password)

	response, err := api.Client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("get %s: server responded with %d", url, response.StatusCode)
		return
	}

	var raw struct {
		Records []struct {
			Nas struct {
				Path string
			}
		}
		Links struct {
			Next struct {
				Href string
			}
		} `json:"_links"`
	}

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&raw)
	if err != nil {
		return
	}

	if raw.Records != nil {
		result.Paths = make([]string, 0, len(raw.Records))
		for _, r := range raw.Records {
			path := r.Nas.Path
			if path != "" {
				result.Paths = append(result.Paths, path)
			}
		}
	}

	nextPage := raw.Links.Next.Href
	if nextPage != "" {
		nextPage, err = resolveURL(url, nextPage)
	}
	result.NextPage = nextPage

	return
}

func (api *API) FetchAll() (paths []string, err error) {
	page, err := api.FetchPage()
	for {
		paths = append(paths, page.Paths...)
		if err != nil {
			return
		}
		if page.NextPage == "" {
			return
		}
		page, err = api.FetchNextPage(page.NextPage)
	}
}
