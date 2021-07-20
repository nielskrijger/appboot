package validate

import "testing"

func BenchmarkFields_Valid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fakeValidUser.Validate()
	}
}

func BenchmarkFields_Invalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fakeInvalidUser.Validate()
	}
}

var benchmarkUser = &fakeUser{
	Name: "John Doe",
}

func BenchmarkStruct_Valid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Struct(benchmarkUser)
	}
}

var benchmarkUserInvalid = &fakeUser{
	Name: "John Do$",
}

func BenchmarkStruct_Invalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Struct(benchmarkUserInvalid)
	}
}
