package main

import (
	"fmt"
	git "github.com/purpleclay/gitz"
)

func main() {
	//cmd.Execute()

	gc, err := git.NewClient()
	if err != nil {
		return
	}

	tags, _ := gc.Tags(
		git.WithShellGlob("*.*.*"),
		git.WithSortBy(git.CreatorDateDesc, git.VersionDesc),
		git.WithCount(1),
	)

	if len(tags) == 1 {
		fmt.Print(tags[0])
	}
}
