package utils

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

func CopyFile(oldFile, newFile string, newFileMode os.FileMode) error {
	oldf, err := os.Open(oldFile)
	if err != nil {
		return errors.WithStack(err)
	}
	defer oldf.Close()

	newf, err := os.OpenFile(newFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, newFileMode)
	if err != nil {
		return errors.WithStack(err)
	}
	defer newf.Close()

	_, err = io.Copy(oldf, newf)

	return errors.WithStack(err)
}
