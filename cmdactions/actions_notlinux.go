// +build !linux

package cmdactions

var (
	ACTIONS map[string]CmdLineAction = map[string]CmdLineAction{
		"help":         &ActionHelp{},
		"init":         &ActionInit{},
		"authenticate": &ActionAuthenticate{},
		"index":        &ActionIndex{},
		"download":     &ActionDownload{},
		"httpd":        &ActionHttpd{}}
)
