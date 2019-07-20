package iter

// Count enumerates an iterator, consuming it and returning the length.
func Count(iterator Iter) (int, error) {
	count := 0
	var err error
	for _, err = iterator.Next(); err == nil; _, err = iterator.Next() {
		count++
	}
	if err == ErrEndIteration {
		return 0, err
	}
	return count, nil
}

// Collect enumerates an iterator, consuming it and returning a slice of the values.
func Collect(iterator Iter) ([]interface{}, error) {
	var values []interface{}
	var value interface{}
	var err error
	for value, err = iterator.Next(); err == nil; value, err = iterator.Next() {
		values = append(values, value)
	}
	if err == ErrEndIteration {
		return nil, err
	}
	return values, nil
}
