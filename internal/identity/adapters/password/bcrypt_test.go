package password

import "testing"

func TestBcryptHashDoesNotStorePlaintextAndVerifies(t *testing.T) {
	passwords := Bcrypt{}
	hash, err := passwords.Hash("a-realistic-test-password")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "a-realistic-test-password" || !passwords.Verify(hash, "a-realistic-test-password") || passwords.Verify(hash, "incorrect") {
		t.Fatal("bcrypt verification invalid")
	}
}
