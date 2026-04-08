package shared

type Metadata map[string]string

func CloneMetadata(input map[string]string) Metadata {
	if input == nil {
		return nil
	}
	out := make(Metadata, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}
