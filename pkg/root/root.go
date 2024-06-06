package root

import (
	"fmt"
)

type Options struct {
	KubeConfig string   `json:"kubeConfig"`
	Args       []string `json:"arg"`
}

func (op *Options) Print() {
	content := `
 __                                 _____                     
_/  |_____________    ____   _______/ ____\___________  _____  
\   __\_  __ \__  \  /    \ /  ___/\   __\/  _ \_  __ \/     \ 
 |  |  |  | \// __ \|   |  \\___ \  |  | (  <_> )  | \/  Y Y  \
 |__|  |__|  (____  /___|  /____  > |__|  \____/|__|  |__|_|  /
                  \/     \/     \/                          \/

The default working directory is /transform. 
Other working directories can be specified using the environment variable TRA_WORKSPACE.
`
	fmt.Print(content)
}
