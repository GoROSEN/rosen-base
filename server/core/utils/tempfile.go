package utils

import (
	"fmt"
	"os"

	uuid "github.com/satori/go.uuid"
)

func TempFilePathName() string {
	return fmt.Sprintf("%v/%v", os.TempDir(), uuid.NewV4().String())
}
