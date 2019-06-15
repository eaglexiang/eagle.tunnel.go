package slice

// EqualStringSlice 比较两个[]string是否相等
func EqualStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, e := range a {
		if e != b[i] {
			return false
		}
	}

	return true
}

// RemoveFromStringSlice 从string slice中删除元素
func RemoveFromStringSlice(dst string, src []string) (result []string) {
	for _, str := range src {
		if str != dst {
			result = append(result, str)
		}
	}
	return
}
