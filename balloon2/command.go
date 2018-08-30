package balloon2

import (
	"encoding/json"
)

// commandType are commands that affect the state of the cluster,
// and must go through raft.
type commandType int

const (
	insert commandType = iota // Commands which modify the database.
	//query                     // Commands which query the database.
)

type command struct {
	Type commandType     `json:"type,omitempty"`
	Sub  json.RawMessage `json:"sub,omitempty"`
}

func newCommand(t commandType, d interface{}) (*command, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return &command{
		Type: t,
		Sub:  b,
	}, nil
}

type insertSubCommand struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
