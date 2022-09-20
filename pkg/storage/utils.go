package storage

import "sort"

func sortFiles(files []File) []File {
	sort.Slice(files, func(i, j int) bool {
		return files[i].CreatedAt.UnixNano() < files[j].CreatedAt.UnixNano()
	})

	return files
}
