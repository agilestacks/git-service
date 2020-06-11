package s3

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"strings"
)

func Unarchive(dir string, archive string) error {
	stream, err := openUrl(archive)
	if err != nil {
		return err
	}
	defer stream.Close()
	var uncompressed io.Reader
	if strings.HasSuffix(archive, "gz") {
		uncompressed, err = gzip.NewReader(stream)
		if err != nil {
			return err
		}
	} else {
		uncompressed = bzip2.NewReader(stream)
	}
	return untar(tar.NewReader(uncompressed), dir)
}
