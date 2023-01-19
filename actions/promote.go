package actions

import (
	"fmt"
	"strings"

	"github.com/ipfs/kuboreleaser/github"
	"github.com/ipfs/kuboreleaser/matrix"
	"github.com/ipfs/kuboreleaser/util"
)

type Promote struct {
	github          *github.Client
	matrix          *matrix.Client
	owner           string
	repo            string
	issueTitle      string
	issueComment    string
	postTitle       string
	postCategory    string
	postTags        []string
	postBody        string
	matrixRoomAlias string
	matrixBotAlias  string
}

func NewPromote(github *github.Client, matrix *matrix.Client, version *util.Version) (*Promote, error) {
	issueComment := fmt.Sprintf(`Early testers ping for %s testing 😄.

- [ ] pacman.store (@RubenKelevra)
- [ ] Infura (@MichaelMure)
- [ ] Textile (@sanderpick)
- [ ] Pinata (@obo20)
- [ ] RTrade (@postables)
- [ ] QRI (@b5)
- [ ] Siderus (@koalalorenzo)
- [ ] Charity Engine (@rytiss, @tristanolive)
- [ ] Fission (@bmann)
- [ ] OrbitDB (@aphelionz)

You're getting this message because you're listed [here](https://github.com/ipfs/kubo/blob/master/docs/EARLY_TESTERS.md#who-has-signed-up). Please update this list if you no longer want to be included.`, version.Version)

	postBody := fmt.Sprintf(`## Kubo %s is out!

See:
- Code: https://github.com/ipfs/kubo/releases/tag/%s
- Binaries: https://dist.ipfs.tech/kubo/%s/
- Docker: `+"`docker pull ipfs/kubo:%s`"+`
- Release Notes (WIP): https://github.com/ipfs/kubo/blob/release-%s/docs/changelogs/%s.md`, version.Version, version.Version, version.Version, version.Version, version.MajorMinor(), version.MajorMinor())

	return &Promote{
		github:          github,
		matrix:          matrix,
		owner:           "ipfs",
		repo:            "kubo",
		issueTitle:      fmt.Sprintf("Release %s", version.MajorMinor()[1:]),
		issueComment:    issueComment,
		postTitle:       fmt.Sprintf("Kubo %s is out!", version.MajorMinor()[1:]),
		postCategory:    "News",
		postTags:        []string{"kubo", "go-ipfs"},
		postBody:        postBody,
		matrixRoomAlias: "#ipfs-chatter:ipfs.io",
		matrixBotAlias:  "@ipfsbot:matrix.org",
	}, nil
}

func (ctx Promote) Check() error {
	issue, err := ctx.github.GetIssue(ctx.owner, ctx.repo, ctx.issueTitle)
	if err != nil {
		return err
	}

	if issue == nil {
		return &util.CheckError{
			Action: util.CheckErrorFail,
			Err:    fmt.Errorf("issue %s not found", ctx.issueTitle),
		}
	}

	comment, err := ctx.github.GetIssueComment(ctx.owner, ctx.repo, issue.GetNumber(), ctx.issueComment)
	if err != nil {
		return err
	}

	if comment == nil {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("comment %s not found", ctx.issueComment)}
	}

	messages, err := ctx.matrix.GetLatestMessagesBy(ctx.matrixRoomAlias, ctx.matrixBotAlias, 100)
	if err != nil {
		return err
	}

	var found bool
	for _, message := range messages {
		body, ok := message.Body()
		if ok && strings.Contains(body, ctx.postTitle) {
			found = true
			break
		}
	}

	if !found {
		return &util.CheckError{Action: util.CheckErrorRetry, Err: fmt.Errorf("post %s not found", ctx.postTitle)}
	}

	return nil
}

func (ctx Promote) Run() error {
	issue, err := ctx.github.GetIssue(ctx.owner, ctx.repo, ctx.issueTitle)
	if err != nil {
		return err
	}

	if issue == nil {
		return fmt.Errorf("issue %s not found", ctx.issueTitle)
	}

	_, err = ctx.github.GetOrCreateIssueComment(ctx.owner, ctx.repo, issue.GetNumber(), ctx.issueComment)
	if err != nil {
		return err
	}

	var confirmation string

	fmt.Printf(`
IPFS Discourse does not have API access enabled.

Please go to https://discuss.ipfs.io and create a new topic with the following content:
Title: %s
Category: %s
Tags: %s
Body: %s

Once you have created the topic, please enter 'yes' to confirm.
Only 'yes' will be accepted to approve.

Enter a value:`, ctx.postTitle, ctx.postCategory, ctx.postTags, ctx.postBody)

	fmt.Scanln(&confirmation)

	if confirmation != "yes" {
		return fmt.Errorf("confirmation is not 'yes'")
	}
	return nil
}
