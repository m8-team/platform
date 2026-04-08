package shared

type ETag string

func (e ETag) String() string {
	return string(e)
}
