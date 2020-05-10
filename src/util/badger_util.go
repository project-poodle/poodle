package util

import (
    "fmt"
    //"time"
    "github.com/dgraph-io/badger"
)


var badger_db_inode *badger.DB

var badger_db_container *badger.DB



func GetBadgerDB_iNode() (*badger.DB, error) {

    if badger_db_inode != nil {
        return badger_db_inode, nil
    }

    opts := badger.DefaultOptions
    opts.Dir = lc_get_badger_db_inode_dir()
    opts.ValueDir = lc_get_badger_db_inode_dir()
    db, err := badger.Open(opts)
    if err != nil {
      return nil, err
    }

    badger_db_inode = db
    return badger_db_inode, nil
}


func GetBadgerDB_Container() (*badger.DB, error) {

    if badger_db_container != nil {
        return badger_db_container, nil
    }

    opts := badger.DefaultOptions
    opts.Dir = lc_get_badger_db_container_dir()
    opts.ValueDir = lc_get_badger_db_container_dir()
    db, err := badger.Open(opts)
    if err != nil {
      return nil, err
    }

    badger_db_container = db
    return badger_db_container, nil
}


func BadgerPut(db *badger.DB, key, value []byte) error {

    err := db.Update(func(txn *badger.Txn) error {
        put_err := txn.Set(key, value)
        return put_err
    })

    if err != nil {
        return err
    }

    return nil
}

func BadgerGet(db *badger.DB, key []byte) ([]byte, error) {
    
    var data []byte = nil
    err := db.View(func(txn *badger.Txn) error {
        data, err := txn.Get(key)
        if err != nil {
            return err
        }
        if data == nil {
            return fmt.Errorf("failed retrieve badger key: %s", key)
        }
        return nil
    })

    if err != nil {
        return nil, err
    }

    return data, nil
}

func BadgerPut_iNode(key, value []byte) error {

    db, err := GetBadgerDB_iNode()
    if err != nil {
        return err
    }

    return BadgerPut(db, key, value)
}

func BadgerGet_iNode(key []byte) ([]byte, error) {

    db, err := GetBadgerDB_iNode()
    if err != nil {
        return nil, err
    }

    return BadgerGet(db, key)
}

func BadgerPut_Container(key, value []byte) error {

    db, err := GetBadgerDB_Container()
    if err != nil {
        return err
    }

    return BadgerPut(db, key, value)
}

func BadgerGet_Container(key []byte) ([]byte, error) {

    db, err := GetBadgerDB_Container()
    if err != nil {
        return nil, err
    }

    return BadgerGet(db, key)
}

