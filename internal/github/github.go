package github

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/google/go-github/v41/github"
	"golang.org/x/sync/errgroup"
)

const (
	githubMaxPerPage = 100
)

type Client struct {
	client *github.Client
}

type RepositoryRelease struct {
	Name    string
	Version string
	Body    string
}

type RepositoryReleaseResponse struct {
	Owner   string
	Name    string
	Version string
	Body    string
	URL     string
	Error   error
}

func (rrr *RepositoryReleaseResponse) IsError() bool {
	return rrr.Error != nil
}

func (gh *Client) CreateReleases(ctx context.Context, owner string, releases []*RepositoryRelease) []*RepositoryReleaseResponse {
	c := make(chan *RepositoryReleaseResponse, len(releases))

	var wg sync.WaitGroup
	wg.Add(len(releases))

	for _, release := range releases {
		release := release

		go func() {
			defer wg.Done()

			r, err := gh.createRelease(ctx, owner, release.Name, release.Version, release.Body)
			response := &RepositoryReleaseResponse{Owner: owner, Name: release.Name, Body: release.Body, Version: release.Version, Error: err}

			if err == nil {
				response.URL = r.GetHTMLURL()
			}

			c <- response
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	var releaseResponses []*RepositoryReleaseResponse
	for release := range c {
		releaseResponses = append(releaseResponses, release)
	}

	return releaseResponses
}

// createRelease Creates a GitHub release provided the owner/repo version and body where the name of the release and the tag will be version.
func (gh *Client) createRelease(ctx context.Context, owner, repo, version, body string) (*github.RepositoryRelease, error) {
	release := &github.RepositoryRelease{
		TagName: github.String(version),
		Body:    github.String(body),
		Name:    github.String(version),
	}

	releaseResponse, r, err := gh.client.Repositories.CreateRelease(ctx, owner, repo, release)
	if err != nil {
		return nil, err
	}

	if err := github.CheckResponse(r.Response); err != nil {
		return nil, err
	}

	return releaseResponse, nil
}

func (gh *Client) tags(ctx context.Context, owner, repo string) ([]*github.RepositoryTag, error) {
	next := 1

	var tags []*github.RepositoryTag
	for {
		options := &github.ListOptions{Page: next, PerPage: githubMaxPerPage}

		responseTags, r, err := gh.client.Repositories.ListTags(ctx, owner, repo, options)
		if err != nil {
			return nil, err
		}

		if err := github.CheckResponse(r.Response); err != nil {
			return nil, err
		}

		tags = append(tags, responseTags...)

		next = r.NextPage
		if next == 0 {
			break
		}
	}

	return tags, nil
}

// mostRecentTagAndChanges get the most recent commits and determine most recent tag for a branch, if branch is not provided the default branch will be used.
func (gh *Client) mostRecentTagAndChanges(ctx context.Context, owner, repo string, tags []*github.RepositoryTag, branch string) (string, *github.RepositoryTag, []*github.RepositoryCommit, error) {
	next := 1

	var commitsSince []*github.RepositoryCommit
	for {
		options := &github.CommitsListOptions{SHA: branch, ListOptions: github.ListOptions{PerPage: githubMaxPerPage, Page: next}}

		commits, r, err := gh.client.Repositories.ListCommits(ctx, owner, repo, options)
		if r.Response.StatusCode == http.StatusNotFound {
			branch = ""
			continue
		}

		if err != nil {
			return "", nil, nil, err
		}

		if err := github.CheckResponse(r.Response); err != nil {
			return "", nil, nil, err
		}

		for _, commit := range commits {
			for _, tag := range tags {
				// find most recent tag associated to this branch
				if commit.GetSHA() == tag.GetCommit().GetSHA() {
					return branch, tag, commitsSince, nil
				}
			}

			commitsSince = append(commitsSince, commit)
		}

		next = r.NextPage
		if next == 0 {
			break
		}
	}

	return branch, nil, commitsSince, nil
}

func (gh *Client) branches(ctx context.Context, owner, repo string) ([]*github.Branch, error) {
	next := 1

	var branches []*github.Branch

	for {
		options := &github.BranchListOptions{ListOptions: github.ListOptions{PerPage: githubMaxPerPage, Page: next}}
		branchesList, r, err := gh.client.Repositories.ListBranches(ctx, owner, repo, options)
		if err != nil {
			return nil, err
		}

		if err := github.CheckResponse(r.Response); err != nil {
			return nil, err
		}

		branches = append(branches, branchesList...)

		next = r.NextPage
		if next == 0 {
			break
		}
	}

	return branches, nil
}

type ReleaseableRepoResponse struct {
	// Total determines when all user Organizations are finished being loaded.
	Total int32
	// Repos in a GitHub organization that have been updated more recently than their newest tag/release.
	Commits   []*github.RepositoryCommit
	Repo      *github.Repository
	LatestTag *github.RepositoryTag
	Branches  []*github.Branch
	Branch    string
}

func (gh *Client) ReleaseableRepo(ctx context.Context, org string, repo *github.Repository, branch string) (*ReleaseableRepoResponse, error) {
	if repo.GetIsTemplate() {
		return nil, nil
	}

	if repo.GetArchived() {
		return nil, nil
	}

	if repo.GetDefaultBranch() == "" {
		return nil, nil
	}

	tags, err := gh.tags(ctx, org, repo.GetName())
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tag for %s/%s: %v", org, repo.GetName(), err)
	}

	branch, tag, commits, err := gh.mostRecentTagAndChanges(ctx, org, repo.GetName(), tags, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits on branch %s for %s/%s: %v", branch, org, repo.GetName(), err)
	}

	if len(commits) == 0 {
		return nil, nil
	}

	branches, err := gh.branches(ctx, org, repo.GetName())
	if err != nil {
		return nil, fmt.Errorf("failed to get branches for %s/%s: %v", org, repo.GetName(), err)
	}

	if branch == "" {
		branch = repo.GetDefaultBranch()
	}

	return &ReleaseableRepoResponse{
		Commits:   commits,
		LatestTag: tag,
		Repo:      repo,
		Branches:  branches,
		Branch:    branch,
	}, nil
}

// ReleaseableReposByOrg async retrival of GitHub repositories for an organization. Returns a channel to listen to for ReleasableRepos
// and the function to run as a goroutine to acquire them.
//
// Filters out: archived, templates, repositories with 0 new commits since last Tag
func (gh *Client) ReleasableReposByOrg(ctx context.Context, org, branch string) (<-chan *ReleaseableRepoResponse, func() error) {
	c := make(chan *ReleaseableRepoResponse)
	next := 1
	var total int32

	errGrp, ctx := errgroup.WithContext(ctx)

	return c, func() error {
		defer close(c)

		for {
			options := &github.RepositoryListByOrgOptions{Sort: "updated", ListOptions: github.ListOptions{PerPage: githubMaxPerPage, Page: next}}

			possibleRepos, r, err := gh.client.Repositories.ListByOrg(ctx, org, options)
			if err != nil {
				return err
			}

			if err := github.CheckResponse(r.Response); err != nil {
				return err
			}

			if len(possibleRepos) == 0 {
				continue
			}

			atomic.AddInt32(&total, int32(len(possibleRepos)))

			for _, repo := range possibleRepos {
				repo := repo

				errGrp.Go(func() error {
					releaseableRepo, err := gh.ReleaseableRepo(ctx, org, repo, branch)
					if err != nil {
						return err
					}

					if releaseableRepo == nil {
						atomic.AddInt32(&total, -1)
						return nil
					}

					releaseableRepo.Total = total
					c <- releaseableRepo
					return nil
				})
			}

			next = r.NextPage
			if next == 0 {
				break
			}
		}

		if err := errGrp.Wait(); err != nil {
			return err
		}

		return nil
	}
}

type OrgResponse struct {
	// Total determines when all user Organizations are finished being loaded.
	Total int
	// Org GitHub organization.
	Org *github.Organization
}

// Orgs async retrival of GitHub organizations. Returns a channel to listen to for OrgsResponse
// and the function to run as a goroutine to acquire them.
func (gh *Client) Orgs(ctx context.Context) (<-chan *OrgResponse, func() error) {
	c := make(chan *OrgResponse)
	next := 1

	errGrp, ctx := errgroup.WithContext(ctx)

	return c, func() error {
		defer close(c)

		for {
			options := &github.ListOptions{Page: next, PerPage: githubMaxPerPage}

			orgs, r, err := gh.client.Organizations.List(ctx, "", options)
			if err != nil {
				return fmt.Errorf("failed to get organizations for authenticated user: %v", err)
			}

			if err := github.CheckResponse(r.Response); err != nil {
				return fmt.Errorf("failed to get organizations for authenticated user: %v", err)
			}

			for _, org := range orgs {
				org := org

				errGrp.Go(func() error {
					orgInfo, r, err := gh.client.Organizations.Get(ctx, org.GetLogin())
					if err != nil {
						return fmt.Errorf("failed to get additional information for organization %s: %v", org.GetLogin(), err)
					}

					if err := github.CheckResponse(r.Response); err != nil {
						return fmt.Errorf("failed to get additional information for organization %s: %v", org.GetLogin(), err)
					}

					c <- &OrgResponse{Total: len(orgs), Org: orgInfo}

					return nil
				})
			}

			next = r.NextPage
			if next == 0 {
				break
			}
		}

		if err := errGrp.Wait(); err != nil {
			return err
		}

		return nil
	}
}

// Username get the username for the currently authenticated user.
func (gh *Client) Username(ctx context.Context) (string, error) {
	user, r, err := gh.client.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %v", err)
	}

	if err := github.CheckResponse(r.Response); err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %v", err)
	}

	return user.GetLogin(), nil
}
