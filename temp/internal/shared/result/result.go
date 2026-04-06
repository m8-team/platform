package result

type Page[T any] struct {
	Items         []T
	NextPageToken string
}
