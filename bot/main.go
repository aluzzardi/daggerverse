package main

import (
	"context"
	"strings"

	"github.com/google/go-github/v59/github"
)

type GithubCi struct{}

func (m *GithubCi) Handle(ctx context.Context, githubToken *Secret, eventName string, eventFile *File) error {
	eventData, err := eventFile.Contents(ctx)
	if err != nil {
		return err
	}
	payload, err := github.ParseWebHook(eventName, []byte(eventData))
	if err != nil {
		return err
	}

	switch ev := payload.(type) {
	case *github.IssueCommentEvent:
		switch ev.GetAction() {
		case "created":
			switch {
			case strings.HasPrefix(ev.Comment.GetBody(), "!echo "):
				comment := dag.GithubComment(
					githubToken,
					ev.GetRepo().GetOwner().GetLogin(),
					ev.GetRepo().GetName(),
					GithubCommentOpts{
						Issue: ev.Issue.GetNumber(),
					},
				)
				if _, err := comment.Create(ctx, strings.TrimPrefix(ev.Comment.GetBody(), "!echo ")); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
