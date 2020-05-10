package util

import (
    "fmt"
    "time"
    "github.com/boltdb/bolt"
)


var bolt_db *bolt.DB



func GetBoltDB() (*bolt.DB, error) {

    if bolt_db != nil {
        return bolt_db, nil
    }

    //fmt.Println("Open DB : " + lc_get_bolt_db_file())
    db, err := bolt.Open(lc_get_bolt_db_file(), 0600, &bolt.Options{Timeout: 1 * time.Second})
    if err != nil {
        return nil, err
    }

    bolt_db = db
    return bolt_db, nil
}


func BoltPut(bucket, key, value []byte) error {

    db, err := GetBoltDB()
    if err != nil {
        return err
    }

    err = db.Update(func(tx *bolt.Tx) error {
        bkt, bkt_err := tx.CreateBucketIfNotExists(bucket)
        if bkt_err != nil {
            return bkt_err
        }
        //b, err := tx.Bucket([]byte("MyBucket"))
        put_err := bkt.Put(key, value)
        return put_err
    })

    if err != nil {
        return err
    }

    return nil
}

func BoltGet(bucket, key []byte) ([]byte, error) {

    db, err := GetBoltDB()
    if err != nil {
        return nil, err
    }

    var data []byte = nil
    err = db.View(func(tx *bolt.Tx) error {
        bkt := tx.Bucket(bucket)
        if bkt == nil {
            return fmt.Errorf("failed retrieve bolt bucket: %s", bucket)
        }
        data = bkt.Get([]byte(key))
        if data == nil {
            return fmt.Errorf("failed retrieve bolt key: %s", key)
        }
        return nil
    })

    if err != nil {
        return nil, err
    }

    return data, nil
}


func BoltList(bucket []byte) (map[string]string, error) {

    db, err := GetBoltDB()
    if err != nil {
        return nil, err
    }

    result := make(map[string]string)

    err = db.View(func(tx *bolt.Tx) error {
        // Assume bucket exists and has keys
        bkt := tx.Bucket(bucket)
        if bkt == nil {
            return nil
        }

        cur := bkt.Cursor()

        for k, v := cur.First(); k != nil; k, v = cur.Next() {
            result[string(k)] = string(v)
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    return result, nil
}
