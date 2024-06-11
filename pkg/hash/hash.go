package hash

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log/slog"
)

func Hash(x interface{}) string {
	hash := sha1.New()
	b, err := json.Marshal(x)
	if err != nil {
		slog.Error("hash %v error, %s", x, err)
		return ""
	}
	hash.Write(b)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
