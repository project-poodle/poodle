package util

import (
    //"bufio"
    //"fmt"
    //"io"
    //"io/ioutil"
    "os"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func Open(filename string) (*os.File) {
    f, err := os.OpenFile("/Users/zhenchen/"+filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0755)
    check(err)

    return f
}

func Write(f *os.File, size int) {

    b1 := make([]byte, size)
    //fmt.Printf("bytes: %x\n", b1)
    _, err := f.Write(b1)
    check(err)
    //fmt.Printf("%d bytes: %s\n", n1, string(b1[:n1]))
}

func Sync(f *os.File) {
    err := f.Sync()
    check(err)
}

func Close(f *os.File) {
    f.Close()
}
