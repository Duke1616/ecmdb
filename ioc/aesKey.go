package ioc

import (
	"github.com/spf13/viper"
)

func AesKey() string {
	return viper.Get("crypto_aes_key").(string)
}
