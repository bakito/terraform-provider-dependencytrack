package provider

import (
	dtrack "github.com/DependencyTrack/client-go"
)

func fetchAllMappedByUI[T any](
	pageFetchFunc func(po dtrack.PageOptions) (dtrack.Page[T], error),
	idOf func(it T) string,
) (map[string]T, error) {
	res, err := dtrack.FetchAll(pageFetchFunc)
	if err != nil {
		return nil, err
	}
	m := mapByID[T](res, idOf)
	return m, nil
}

func mapByID[T any](res []T, idOf func(it T) string) map[string]T {
	m := make(map[string]T)
	for i := range res {
		item := res[i]
		m[idOf(item)] = item
	}
	return m
}
