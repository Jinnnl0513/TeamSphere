package ws

func ptrStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
