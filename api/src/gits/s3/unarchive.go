package s3

import (
	"archive/tar"
	"compress/bzip2"
	"io"
)

func Unarchive(dir string, archive string) error {
	stream, err := openArchive(archive)
	if err != nil {
		return err
	}
	defer stream.Close()
	return unarchiveStream(dir, stream)
}

func unarchiveStream(dir string, stream io.Reader) error {
	bz2 := bzip2.NewReader(stream)
	return untar(tar.NewReader(bz2), dir)
}
