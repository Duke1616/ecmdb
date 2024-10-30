package main

import (
	"fmt"
	"github.com/Duke1616/ecmdb/cmd"
	"github.com/fatih/color"
	git "github.com/purpleclay/gitz"
)

var (
	version string
)

func main() {
	ver := version
	if version == "" {
		ver = latestTag()
	}

	fmt.Println("  ______    _____   __  __   _____    ____  ")
	fmt.Println(" |  ____|  / ____| |  \\/  | |  __ \\  |  _ \\ ")
	fmt.Println(" | |__    | |      | \\  / | | |  | | | |_) |")
	fmt.Println(" |  __|   | |      | |\\/| | | |  | | |  _ < ")
	fmt.Println(" | |____  | |____  | |  | | | |__| | | |_) |")
	fmt.Println(" |______|  \\_____| |_|  |_| |_____/  |____/ ")

	// 使用颜色来突出显示
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf(" %s: %s\n", cyan("Service Version"), green(ver))

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
