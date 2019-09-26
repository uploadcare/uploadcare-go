package uploadcare

// String returns a pointer to the string value passed in
func String(v string) *string { return &v }

// Int returns a pointer to the int value passed in
func Int(v int) *int { return &v }
