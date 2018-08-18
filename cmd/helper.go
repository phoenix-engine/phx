package cmd

// Must wraps a Flag getter which has a known-good state.
func Must(s string, e error) string {
	if e != nil {
		panic(e)
	}
	return s
}
