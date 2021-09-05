package set

import (
	"sort"
)

func SortDeDup(l []string) []string {
	n := len(l)
	if n <= 1 {
		return l
	}
	sort.Strings(l)

	j := 1
	for i := 1; i < n; i++ {
		if l[i] != l[i-1] {
			l[j] = l[i]
			j++
		}
	}

	return l[0:j]
}
