package ports

type FilterValidator interface {
	Validate(raw string) error
}

type OrderValidator interface {
	Validate(raw string) error
}
