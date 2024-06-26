package remote

import (
	"strings"
)

type Commands []string

type File struct {
	Src string `json:"src,omitempty"`
	Dst string `json:"dst,omitempty"`
}

type Command struct {
	Cmds   Commands `json:"cmds,omitempty"`
	FileUp []File   `json:"fileUp,omitempty"`
}

func (c Command) String() string {
	return strings.Join(c.Cmds, " && ")
}

func (c Command) List() []string {
	return c.Cmds
}
