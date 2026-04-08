package ports

type UUIDGenerator interface {
	NewString() string
}
