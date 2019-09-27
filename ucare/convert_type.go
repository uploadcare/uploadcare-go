package ucare

// String returns a pointer to the string value passed in
func String(v string) *string { return &v }

// StringVal returns the value of the string pointer passed in or
// "" if the pointer is nil.
func StringVal(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// Int64 returns a pointer to the int value passed in
func Int64(v int64) *int64 { return &v }

// Int64Val returns the value of the int pointer passed in or
// 0 if the pointer is nil.
func Int64Val(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
}

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// BoolVal returns the value of the bool pointer passed in or
// false if the pointer is nil.
func BoolVal(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}
