package uuid

import googleuuid "github.com/google/uuid"

type Generator struct{}

func (Generator) NewString() string {
	return googleuuid.NewString()
}
