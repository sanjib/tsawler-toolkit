package toolkit

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type that allows access to its various utility methods.
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

// RandomString generates a base32 random string of n size.
func RandomString(n int) string {
	bb := make([]byte, n)
	_, err := rand.Read(bb)
	if err != nil {
		log.Println(err)
		return ""
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bb)[:n]
}

// RandomString generates a random string from randomStringSource of n size.
func (t *Tools) RandomString(n int) string {
	dst, src := make([]rune, n), []rune(randomStringSource)
	for i := range dst {
		p, _ := rand.Prime(rand.Reader, len(src))
		x, y := p.Uint64(), uint64(len(src))
		//fmt.Println("x, y, x%y:", x, y, x%y)
		dst[i] = src[x%y]
	}
	return string(dst)
}

// UploadedFile save information about the file to send back to user.
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}
	var uploadedFiles []UploadedFile

	// use default MaxFileSize
	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}
	if err := r.ParseMultipartForm(int64(t.MaxFileSize)); err != nil {
		return nil, err
	}

	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			var err error
			uploadedFiles, err = func(uploadedFiles []UploadedFile) ([]UploadedFile, error) {
				var uploadedFile UploadedFile
				srcFile, err := fileHeader.Open()
				if err != nil {
					return nil, err
				}
				defer srcFile.Close()

				// Figure out the file type
				bb := make([]byte, 512)
				if _, err := srcFile.Read(bb); err != nil {
					return nil, err
				}
				filetype := http.DetectContentType(bb)

				// Is file type permitted?
				allowed := false
				if len(t.AllowedFileTypes) > 0 {
					for _, allowedType := range t.AllowedFileTypes {
						if strings.EqualFold(filetype, allowedType) {
							allowed = true
							break
						}
					}
				} else {
					allowed = true
				}
				if !allowed {
					return nil, errors.New("uploaded file type is not permitted")
				}

				// Copy file
				if _, err := srcFile.Seek(0, 0); err != nil {
					return nil, err
				}
				uploadedFile.NewFileName = fileHeader.Filename
				if renameFile {
					uploadedFile.NewFileName =
						fmt.Sprintf("%s%s", RandomString(8), filepath.Ext(fileHeader.Filename))
				}
				dstFile, err := os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName))
				if err != nil {
					return nil, err
				}
				defer dstFile.Close()

				n, err := io.Copy(dstFile, srcFile)
				if err != nil {
					return nil, err
				}
				uploadedFile.FileSize = n
				uploadedFile.OriginalFileName = fileHeader.Filename

				uploadedFiles = append(uploadedFiles, uploadedFile)
				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	return uploadedFiles, nil
}
