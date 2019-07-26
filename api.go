package main

import "time"

type Repository struct {
	Scm     string `json:"scm"`
	Website string `json:"website"`
	HasWiki bool   `json:"has_wiki"`
	Name    string `json:"name"`
	Links   struct {
		Watchers struct {
			Href string `json:"href"`
		} `json:"watchers"`
		Branches struct {
			Href string `json:"href"`
		} `json:"branches"`
		Tags struct {
			Href string `json:"href"`
		} `json:"tags"`
		Commits struct {
			Href string `json:"href"`
		} `json:"commits"`
		Clone []struct {
			Href string `json:"href"`
			Name string `json:"name"`
		} `json:"clone"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Source struct {
			Href string `json:"href"`
		} `json:"source"`
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
		Hooks struct {
			Href string `json:"href"`
		} `json:"hooks"`
		Forks struct {
			Href string `json:"href"`
		} `json:"forks"`
		Downloads struct {
			Href string `json:"href"`
		} `json:"downloads"`
		Issues struct {
			Href string `json:"href"`
		} `json:"issues"`
		PullRequests struct {
			Href string `json:"href"`
		} `json:"pullrequests"`
	} `json:"links"`
	ForkPolicy string    `json:"fork_policy"`
	UUID       string    `json:"uuid"`
	Language   string    `json:"language"`
	CreatedOn  time.Time `json:"created_on"`
	MainBranch struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"mainbranch"`
	FullName  string `json:"full_name"`
	HasIssues bool   `json:"has_issues"`
	Owner     struct {
		DisplayName string `json:"display_name"`
		UUID        string `json:"uuid"`
		Links       struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
		} `json:"links"`
		Nickname  string `json:"nickname"`
		Type      string `json:"type"`
		AccountID string `json:"account_id"`
	} `json:"owner"`
	UpdatedOn   time.Time `json:"updated_on"`
	Size        int       `json:"size"`
	Type        string    `json:"type"`
	Slug        string    `json:"slug"`
	IsPrivate   bool      `json:"is_private"`
	Description string    `json:"description"`
}

type BitbucketResponse struct {
	Pagelen      int          `json:"pagelen"`
	Size         int          `json:"size"`
	Repositories []Repository `json:"values"`
	Page         int          `json:"page"`
	Next         string       `json:"next"`
}
