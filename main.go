package main

import (
	"fmt"
	"github.com/Duke1616/ecmdb/cmd"
	git "github.com/purpleclay/gitz"
)

var (
	version string
)

const (
	ProjectName    = "MyAwesomeProject"
	ProjectVersion = "1.0.0"
	BuiltWith      = "123"             // 获取Go的版本
	BuiltAt        = "An unknown time" // 这里可以替换为具体的构建时间
	GitCommit      = "unknown commit"  // 这里可以替换为具体的Git提交哈希
)

func main() {
	ver := version
	fmt.Printf("beflore version:  %s \n", ver)
	if version == "" {
		ver = latestTag()
	}

	fmt.Printf("after version:  %s \n", ver)
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
