package main

func IndexOf(list []string, value string) int {
	for i, v := range list {
		if v == value {
			return i
		}
	}

	return -1
}

func Remove(list []string, value string) []string {
	index := IndexOf(list, value)
	if index < 0 {
		return list
	}

	return append(list[:index], list[index+1:]...)
}
