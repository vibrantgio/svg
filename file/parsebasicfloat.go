package file

func parseBasicFloat(s string) (float64, error) {
	value, _, err := parseUnit(s)
	return value, err
}
