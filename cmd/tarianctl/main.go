package main

import (
	ver "github.com/kube-tarian/tarian/cmd"
	"github.com/kube-tarian/tarian/cmd/tarianctl/cmd"
)

var (
	version    = "dev"
	commit     = "main"
	versionStr = version + " (" + commit + ")"
)

func main() {
	ver.SetVersion(versionStr)
	cmd.Execute()
}
