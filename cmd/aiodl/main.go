package main

import "github.com/aiomayo/aiodl/cmd/aiodl/cmd"

var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
