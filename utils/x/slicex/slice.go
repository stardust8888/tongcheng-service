package slicex

import "reflect"

func RemoveDuplicates[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0)

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func ExtractInt[T comparable](slice []T, name string) []int {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return []int{}
	}

	ints := make([]int, 0, s.Len())

	for i := 0; i < s.Len(); i++ {
		elem := s.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			continue
		}
		field := elem.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		if field.Kind() == reflect.Int {
			ints = append(ints, int(field.Int()))
		}
	}

	return ints
}

func ExtractString[T comparable](slice []T, name string) []string {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return []string{}
	}

	strs := make([]string, 0, s.Len())

	for i := 0; i < s.Len(); i++ {
		elem := s.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			continue
		}
		field := elem.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		if field.Kind() == reflect.String {
			strs = append(strs, field.String())
		}
	}

	return strs
}

func FilterInSlice[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func MapInSlice[T, U any](slice []T, mapper func(T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

func ContainInSlice[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func FindInSlice[T any](slice []T, predicate func(T) bool) (T, bool) {
	var zero T
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	return zero, false
}

func CheckInSlice[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if !predicate(item) {
			return false
		}
	}
	return true
}
