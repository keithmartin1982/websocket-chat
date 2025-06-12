package main

import (
	"golang.org/x/crypto/bcrypt"
)

func genHash(pass []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(pass, bcrypt.MinCost)
}

func checkPass(pass, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, pass)
}
