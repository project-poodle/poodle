package util

import (
    //"fmt"
    "time"
    "testing"
)

func setupTimestampData() {
}

func TestSetupTimestamp(t *testing.T) {
    setupTimestampData()
}

func BenchmarkUnixNano(b *testing.B) {
    setupTimestampData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        time.Now().UnixNano()
    }
}

