package main

import (
	"github.com/Duke1616/ecmdb/cmd"
	git "github.com/purpleclay/gitz"
)

var version string

func main() {
	ver := version
	if version == "" {
		ver = latestTag()
	}

	cmd.Execute(ver)
}

func latestTag() string {
	gc, err := git.NewClient()
	if err != nil {
		return ""
	}

	tags, _ := gc.Tags(
		git.WithShellGlob("*.*.*"),
		git.WithSortBy(git.CreatorDateDesc, git.VersionDesc),
		git.WithCount(1),
	)

	return tags[0]
}
