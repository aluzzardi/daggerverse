package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v59/github"
)

type GithubComment struct {
	GithubToken *Secret // +private
	MessageID   string  // +private
	Owner       string  // +private
	Repo        string  // +private
	Issue       int     // +private
	Commit      string  // +private
}

func New(
	ctx context.Context,
	// Github API token
	githubToken *Secret,
	// A stable identifier to enable editing the same comment in-place.
	// The key is included in the comment message but invisible
	// +optional
	// +default="github.com/aluzzardi/daggerverse/github-comment"
	messageID string,
	// The github repository
	// Supported formats:
	// - github.com/dagger/dagger
	// - dagger/dagger
	// - https://github.com/dagger/dagger
	// - https://github.com/dagger/dagger.git
	repo string,
	// Comment on the given github issue
	// +optional
	issue int,
	// Comment on the given commit
	// +optional
	commit string,
) (*GithubComment, error) {
	// Strip .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")

	// Remove https:// or http:// prefix if present
	repo = strings.TrimPrefix(repo, "https://")
	repo = strings.TrimPrefix(repo, "http://")

	// Remove github.com/ prefix if present
	repo = strings.TrimPrefix(repo, "github.com/")

	// Split remaining string into owner/repo
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", repo)
	}

	return &GithubComment{
		GithubToken: githubToken,
		MessageID:   messageID,
		Owner:       parts[0],
		Repo:        parts[1],
		Issue:       issue,
		Commit:      commit,
	}, nil
}

func (m *GithubComment) newClient(ctx context.Context) (*github.Client, error) {
	token, err := m.GithubToken.Plaintext(ctx)
	if err != nil {
		return nil, err
	}
	return github.NewClient(nil).WithAuthToken(token), nil
}

func marker(messageID string) string {
	return fmt.Sprintf("<!-- marker: %s -->", messageID)
}

func (m *GithubComment) markBody(body string) *string {
	marked := marker(m.MessageID) + "\n" + body
	return &marked
}

func (m *GithubComment) getIssueFromCommit(ctx context.Context, ghc *github.Client, commitSha string) (int, error) {
	prs, _, err := ghc.PullRequests.ListPullRequestsWithCommit(ctx, m.Owner, m.Repo, commitSha, nil)
	if err != nil {
		return 0, err
	}
	if len(prs) == 0 {
		return 0, fmt.Errorf("commit %s not found in any pull request", commitSha)
	}
	return prs[0].GetNumber(), nil
}

func (m *GithubComment) findComment(ctx context.Context, ghc *github.Client) (*github.IssueComment, int, error) {
	opt := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}
	var (
		issue = m.Issue
		err   error
	)
	if issue == 0 {
		if m.Commit == "" {
			return nil, 0, fmt.Errorf("either issue or commit must be set")
		}
		issue, err = m.getIssueFromCommit(ctx, ghc, m.Commit)
		if err != nil {
			return nil, 0, err
		}
	}
	for {
		comments, resp, err := ghc.Issues.ListComments(ctx, m.Owner, m.Repo, issue, opt)
		if err != nil {
			return nil, 0, err
		}

		for _, comment := range comments {
			if comment.Body != nil && strings.HasPrefix(*comment.Body, marker(m.MessageID)) {
				return comment, issue, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return nil, issue, nil
}

// Create or update the comment on github
func (m *GithubComment) Create(ctx context.Context, body string) (*string, error) {
	ghc, err := m.newClient(ctx)
	if err != nil {
		return nil, err
	}
	existingComment, issue, err := m.findComment(ctx, ghc)
	if err != nil {
		return nil, err
	}

	var comment *github.IssueComment
	if existingComment != nil {
		existingComment.Body = m.markBody(body)
		comment, _, err = ghc.Issues.EditComment(ctx, m.Owner, m.Repo, *existingComment.ID, existingComment)
	} else {
		comment, _, err = ghc.Issues.CreateComment(ctx, m.Owner, m.Repo, issue, &github.IssueComment{
			Body: m.markBody(body),
		})
	}
	if err != nil {
		return nil, err
	}

	return comment.HTMLURL, nil
}

// Delete the comment on github
func (m *GithubComment) Delete(ctx context.Context) error {
	ghc, err := m.newClient(ctx)
	if err != nil {
		return err
	}

	comment, _, err := m.findComment(ctx, ghc)
	if err != nil {
		return err
	}
	if comment == nil {
		return nil
	}

	_, err = ghc.Issues.DeleteComment(ctx, m.Owner, m.Repo, *comment.ID)
	return err
}

// Add an emoji reaction to the comment
func (m *GithubComment) React(
	ctx context.Context,
	// The kind of reaction.
	// Supported values: "+1", "-1", "laugh", "confused", "heart", "hooray", "rocket", or "eyes".
	kind string,
) error {
	ghc, err := m.newClient(ctx)
	if err != nil {
		return err
	}

	comment, _, err := m.findComment(ctx, ghc)
	if err != nil {
		return err
	}
	if comment == nil {
		return nil
	}

	_, _, err = ghc.Reactions.CreateIssueCommentReaction(ctx, m.Owner, m.Repo, *comment.ID, kind)
	return err
}
