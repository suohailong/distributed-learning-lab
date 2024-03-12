package wal

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestKvStore(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	kv := NewKvStore("testdata/waltest.log")

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		// kv.Put(key, value)
		v := kv.Get(key)
		fmt.Println("value: ", v)
		if v != value {
			t.Fatal("kv store get failed")
		}
	}
	time.Sleep(2 * time.Second)
}
