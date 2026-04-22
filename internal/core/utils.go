package core

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func UUID(separator bool) string {
	id := uuid.New().String()
	if !separator {
		return id
	}
	return id
}

func UnixTimestamp() int64 {
	return time.Now().Unix()
}

func Timestamp() int64 {
