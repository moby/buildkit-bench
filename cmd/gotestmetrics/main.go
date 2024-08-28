package main

import (
	"log"

	"github.com/alecthomas/kong"
)

var cli struct {
	Debug bool `kong:"name='debug',env='DEBUG',default='false',help='Enable debug mode.'"`

	Parse parseCmd `kong:"cmd,help='Parse go test output.'"`
	List  listCmd  `kong:"cmd,help='List tests from configuration.'"`
	Merge mergeCmd `kong:"cmd,help='Merge benchmark results.'"`
}

type Context struct {
	Debug bool
}

func main() {
	log.SetFlags(0)
	ctx := kong.Parse(&cli,
		kong.Name("gotestmetrics"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))
	ctx.FatalIfErrorf(ctx.Run(&Context{Debug: cli.Debug}))
}
