package i18n

func translated(overrides map[string]string) map[string]string {
	result := make(map[string]string, len(english))
	for key, value := range english {
		result[key] = value
	}
	for key, value := range overrides {
		result[key] = value
	}
	return result
}
