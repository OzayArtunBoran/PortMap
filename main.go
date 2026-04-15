package main

import "github.com/ozayartunboran/portmap/cmd"

var (
	version   string
	buildTime string
)

func main() {
	cmd.SetVersionInfo(version, buildTime)
	cmd.Execute()
}
