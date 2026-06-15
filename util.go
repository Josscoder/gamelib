package gamelib

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnZipFile extracts a complete .zip archive to the provided destination directory (dest).
func UnZipFile(srcZip string, dest string) error {
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer r.Close()

	// Clean the destination path for security checks
	dest = filepath.Clean(dest)

	for _, f := range r.File {
		// Zip Slip Protection: assures that the resolved path is fully inside the 'dest' folder
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, dest+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip (Zip Slip detected): %s", f.Name)
		}

		// If the entry is a directory, recreate it
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// If it is a file, assure its parent folder exists
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// Open/create the final file, keeping the permissions that came from the Zip
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		// Open the compressed original file from inside the Zip
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		// Fast binary copy
		_, err = io.Copy(outFile, rc)

		// Close both files immediately. (We avoid throwing a 'defer' inside a loop
		// because that would hold ALL files open in memory until the whole loop finishes)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
