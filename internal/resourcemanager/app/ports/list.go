package ports

type SortDirection string

const (
	SortDirectionAscending  SortDirection = "asc"
	SortDirectionDescending SortDirection = "desc"
)

func (d SortDirection) IsValid() bool {
	return d == SortDirectionAscending || d == SortDirectionDescending
}
