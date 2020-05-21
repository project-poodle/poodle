package util

import (
    "fmt"
    "crypto/rand"
    "crypto/ecdsa"
    "testing"
)


func setupData() ([]byte) {
    data := make([]byte, 4096)
    for i:=0; i<len(data); {
        if n, err := rand.Read(data); err != nil {
            fmt.Printf("%s\n", err)
        } else {
            i += n
        }
    }
    //fmt.Printf("\nData Size : %d\n", len(data))
    return data
}

func BenchmarkSumSHA256(b *testing.B) {
    data := setupData()
    //fmt.Printf("%x\n", data)

    //b.ResetTimer()
    for i := 0; i < b.N; i++ {
        SumSHA256(data)
    }
}

func BenchmarkSumMD5(b *testing.B) {
    data := setupData()
    //fmt.Printf("%x\n", data)

    //b.ResetTimer()
    for i := 0; i < b.N; i++ {
        SumMD5(data)
    }
}


func TestECDSAGenerateKey(t *testing.T) {
    key := ECDSAGenerateKey()
    if key == nil {
        t.Fatal("failed")
    }
}

func BenchmarkECDSAGenerateKey(b *testing.B) {
    for i := 0; i < b.N; i++ {
        ECDSAGenerateKey()
    }
}


func BenchmarkECDSASign(b *testing.B) {
    priv_key := ECDSAGenerateKey()
    block := make([]byte, 4096)
    for i:=0; i<len(block); i++ {
        block[i] = byte(i)
    }
    hash := SumSHA256d(block)
    //b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ECDSASign(priv_key, hash)
    }
}


func BenchmarkECDSAVerify(b *testing.B) {
    priv_key := ECDSAGenerateKey()
    pub_key  := priv_key.Public()
    block := make([]byte, 4096)
    for i:=0; i<len(block); i++ {
        block[i] = byte(i)
    }
    hash := SumSHA256d(block)
    r, s, _ := ECDSASign(priv_key, hash)
    //b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ECDSAVerify(pub_key.(*ecdsa.PublicKey), hash, r, s)
    }
}


func BenchmarkECDSACalculateY(b *testing.B) {
    priv_key := ECDSAGenerateKey()
    pub_key  := priv_key.Public().(*ecdsa.PublicKey)
    //b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ECDSACalculateY(pub_key.Curve, pub_key.X)
    }
}


func BenchmarkECDSAGetPrivateKey(b *testing.B) {
    priv_key := ECDSAGenerateKey()
    priv_bytes := priv_key.D.Bytes()
    //b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ECDSAGetPrivateKey(priv_bytes)
    }
}

func BenchmarkECDSAGetPublicKey(b *testing.B) {
    priv_key := ECDSAGenerateKey()
    pub_key  := priv_key.Public().(*ecdsa.PublicKey)
    pub_bytes := pub_key.X.Bytes()
    //b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ECDSAGetPublicKey(pub_bytes)
    }
}
