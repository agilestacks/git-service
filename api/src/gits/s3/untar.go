package s3

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var dirMode = os.FileMode(0755)

func untar(tr *tar.Reader, dir string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		err = untarEntry(tr, header, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func untarEntry(tr *tar.Reader, header *tar.Header, dir string) error {
	name := filepath.Join(dir, header.Name)
	switch header.Typeflag {
	case tar.TypeDir:
		return mkdir(name)
	case tar.TypeReg, tar.TypeRegA:
		return writeFile(name, tr, header.FileInfo().Mode())
	default:
		log.Printf("Ignoring %q while untaring into %s: unsupported type flag: %c", header.Name, dir, header.Typeflag)
	}
	return nil
}

func mkdir(dir string) error {
	err := os.MkdirAll(dir, dirMode)
	if err != nil {
		return fmt.Errorf("Failed to mkdir %q: %v", dir, err)
	}
	return nil
}

func writeFile(name string, content io.Reader, mode os.FileMode) error {
	dir := filepath.Dir(name)
	err := os.MkdirAll(dir, dirMode)
	if err != nil {
		return fmt.Errorf("Failed to mkdir %q: %v", dir, err)
	}

	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("Failed to create file %q: %v", name, err)
	}
	defer file.Close()

	err = file.Chmod(mode)
	if err != nil {
		log.Printf("Failed to change file %q mode to %v: %v", name, mode, err)
	}

	_, err = io.Copy(file, content)
	if err != nil {
		return fmt.Errorf("Failed to write file %q: %v", name, err)
	}

	return nil
}
