package randoms

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func RandCryptoIntN(max int64) (*big.Int, error) { // [0,max)
	maxLimit := big.NewInt(max)

	secureNum, err := rand.Int(rand.Reader, maxLimit)
	if err != nil {
		return nil, err
	}

	return secureNum, nil

}

func RandCryptoIntNInRange(min, max int64) (int64, error) { // [min,max)
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}

	rangeSize := big.NewInt(max - min + 1)

	n, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return 0, err
	}

	return n.Int64() + min, nil
}

func RandCryptoBytes(length int) ([]byte, error) {
	token := make([]byte, length)

	_, err := rand.Read(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func RandCryptoStringFromSource(length int) (string, error) {
	result := make([]byte, length)
	alphabetLength := big.NewInt(int64(len(alphabet)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, alphabetLength)
		if err != nil {
			return "", err
		}
		result[i] = alphabet[num.Int64()]
	}

	return string(result), nil
}
