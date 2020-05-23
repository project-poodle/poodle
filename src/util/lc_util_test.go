package util

import (
	"crypto/aes"
	"fmt"
	"os"
	"testing"
)

func TestLCKeyGen(t *testing.T) {

	cls := CLS_NODE

	secret := "poodle"
	padded_secret := make([]byte, aes.BlockSize)
	copy(padded_secret, secret)

	priv_key := ECDSAGenerateKey()
	pub_key := priv_key.PublicKey

	id := fmt.Sprintf("%x", pub_key.X.Bytes())
	err := LCSaveKeyPair(cls, id, pub_key.X, priv_key.D, padded_secret)
	if err != nil {
		t.Fatal(err)
	}

	_, err = LCLoadPubKey(cls, id)
	if err != nil {
		t.Fatal(err)
	}

	_, err = LCLoadPrivKey(cls, id, padded_secret)
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(LCGetEtcDir(cls) + "/" + id)
}

func BenchmarkLCKeyGen(b *testing.B) {

	cls := CLS_NODE

	secret := "poodle"
	padded_secret := make([]byte, aes.BlockSize)
	copy(padded_secret, secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		priv_key := ECDSAGenerateKey()
		pub_key := priv_key.PublicKey

		id := fmt.Sprintf("%x", pub_key.X.Bytes())
		err := LCSaveKeyPair(cls, id, pub_key.X, priv_key.D, padded_secret)
		if err != nil {
			b.Fatal(err)
		}

		_, err = LCLoadPubKey(cls, id)
		if err != nil {
			b.Fatal(err)
		}

		_, err = LCLoadPrivKey(cls, id, padded_secret)
		if err != nil {
			b.Fatal(err)
		}

		os.RemoveAll(LCGetEtcDir(cls) + "/" + id)
	}
}
