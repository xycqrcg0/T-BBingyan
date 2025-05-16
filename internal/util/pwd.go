package util

import "golang.org/x/crypto/bcrypt"

func HashPwd(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(b), err
}

func ParsePwd(hashed string, pwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pwd))
}
