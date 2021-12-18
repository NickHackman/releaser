package service

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/google/go-github/v41/github"
	"golang.org/x/sync/errgroup"
)

const (
	githubMaxPerPage = 100
)

type GitHub struct {
	client *github.Client
}

type ReleaseableRepoResponse struct {
	// Total determines when all user Organizations are finished being loaded.
	Total int32
	// Repos in a GitHub organization that have been updated more recently than their newest tag/release.
	Commits   []*github.RepositoryCommit
	Repo      *github.Repository
	LatestTag *github.RepositoryTag
}

// Creates a GitHub release provided the owner/repo version and body where the name of the release and the tag will be version.
func (gh *GitHub) CreateRelease(ctx context.Context, owner, repo, version, body string) (*github.RepositoryRelease, error) {
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

func (gh *GitHub) latestTag(ctx context.Context, owner, repo string) (*github.RepositoryTag, error) {
	options := &github.ListOptions{Page: 1, PerPage: 1}

	tags, r, err := gh.client.Repositories.ListTags(ctx, owner, repo, options)
	if err != nil {
		return nil, err
	}

	if err := github.CheckResponse(r.Response); err != nil {
		return nil, err
	}

	if len(tags) == 0 {
		return nil, nil
	}

	return tags[0], nil
}

func (gh *GitHub) commitsSince(ctx context.Context, owner, repo string, sha string) ([]*github.RepositoryCommit, error) {
	next := 1

	var commitsSince []*github.RepositoryCommit
	for {
		options := &github.CommitsListOptions{ListOptions: github.ListOptions{PerPage: githubMaxPerPage, Page: next}}

		commits, r, err := gh.client.Repositories.ListCommits(ctx, owner, repo, options)
		if err != nil {
			return nil, err
		}

		if err := github.CheckResponse(r.Response); err != nil {
			return nil, err
		}

		for _, commit := range commits {
			if commit.GetSHA() == sha {
				return commitsSince, nil
			}

			commitsSince = append(commitsSince, commit)
		}

		next = r.NextPage
		if next == 0 {
			break
		}
	}

	return commitsSince, nil
}

func (gh *GitHub) releaseableRepo(ctx context.Context, org string, repo *github.Repository) (*ReleaseableRepoResponse, error) {
	if repo.GetIsTemplate() {
		return nil, nil
	}

	if repo.GetArchived() {
		return nil, nil
	}

	tag, err := gh.latestTag(ctx, org, repo.GetName())
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tag for %s/%s: %v", org, repo.GetName(), err)
	}

	var commitsSince string
	if tag != nil {
		commitsSince = tag.GetCommit().GetSHA()
	}

	commits, err := gh.commitsSince(ctx, org, repo.GetName(), commitsSince)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits since %s for %s/%s: %v", commitsSince, org, repo.GetName(), err)
	}

	if len(commits) == 0 {
		return nil, nil
	}

	return &ReleaseableRepoResponse{Commits: commits, LatestTag: tag, Repo: repo}, nil
}

// ReleaseableReposByOrg async retrival of GitHub repositories for an organization. Returns a channel to listen to for ReleasableRepos
// and the function to run as a goroutine to acquire them.
//
// Filters out: archived, templates, repositories with 0 new commits since last Tag
func (gh *GitHub) ReleasableReposByOrg(ctx context.Context, org string) (<-chan *ReleaseableRepoResponse, func() error) {
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
					releaseableRepo, err := gh.releaseableRepo(ctx, org, repo)
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
func (gh *GitHub) Orgs(ctx context.Context) (<-chan *OrgResponse, func() error) {
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
				return fmt.Errorf("failed to get organization for authenticate user: %v", err)
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
