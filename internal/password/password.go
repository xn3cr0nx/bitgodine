package password

import "golang.org/x/crypto/bcrypt"

// Hash returnes the hased input password
func Hash(pass string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(hashed), err
}

// Verify compares the hashed string with plain password
func Verify(hash, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}
