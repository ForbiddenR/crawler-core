package utils

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEncryptAesPassword(t *testing.T) {
	plainText := "crawlab"
	encryptedText, err := EncryptAES(plainText)
	require.Nil(t, err)
	decryptedText, err := DecryptAES(encryptedText)
	require.Nil(t, err)
	fmt.Printf("plainText: %s", plainText)
	fmt.Printf("encryptedText: %s", encryptedText)
	fmt.Printf("decryptedText: %s", decryptedText)
	require.Equal(t, decryptedText, plainText)
	require.NotEqual(t, decryptedText, encryptedText)
}
