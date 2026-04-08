package port

type UUIDGenerator interface {
	NewString() string
}
