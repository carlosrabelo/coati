package templates

import (
	_ "embed"
)

//go:embed hosts.tpl
var HostsTemplate string
