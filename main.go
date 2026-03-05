package main

import (
	_ "wish-fullfilement-fiction/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"wish-fullfilement-fiction/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
