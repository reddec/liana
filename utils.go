package liana

func ToStringSet(s []string) map[string]bool {
	v := make(map[string]bool)
	for _, k := range s {
		v[k] = true
	}
	return v
}
