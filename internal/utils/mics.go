package utils

import (
	crypto "crypto/rand"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"regexp"
)

func PreparePhone(phone string) string {
	var re = regexp.MustCompile(`\D`)
	phone = re.ReplaceAllString(phone, "")

	if len(phone) == 10 {
		return fmt.Sprintf("7%s", phone)
	}
	if phone[:1] == "8" {
		return fmt.Sprintf("7%s", phone[1:])
	}
	return phone
}

func RemoveUUIDIndex(s []*uuid.UUID, index int) []*uuid.UUID {
	return append(s[:index], s[index+1:]...)
}

func NewCryptoRand() int64 {
	safeNum, err := crypto.Int(crypto.Reader, big.NewInt(999999))
	if err != nil {
		panic(err)
	}
	return safeNum.Int64()
}


