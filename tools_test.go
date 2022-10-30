package toolkit_test

import (
	"fmt"
	"github.com/sanjib/tsawler-toolkit"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	randStr := toolkit.RandomString(8)
	t.Run("MatchLen", func(t *testing.T) {
		want := 8
		got := len(randStr)
		if want != got {
			t.Errorf("want %d, got %d", want, got)
		}
	})
}

func TestTools_RandomStringMethod(t *testing.T) {
	tools := toolkit.Tools{}
	randStr := tools.RandomString(8)
	t.Run("MatchLen", func(t *testing.T) {
		want := 8
		got := len(randStr)
		if want != got {
			t.Errorf("want %d, got %d", want, got)
		}
	})
}

func TestTools_UploadFiles(t *testing.T) {
	testData := []struct {
		name         string
		allowedTypes []string
		renameFile   bool
		errExpected  bool
	}{
		{"NoRename", []string{"image/png"}, false, false},
		{"AllowedRename", []string{"image/png"}, true, false},
		{"NotAllowedFileType", []string{"image/jpg"}, false, true},
	}

	var wg sync.WaitGroup

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			// Setup pipe to avoid buffering
			pipeReader, pipeWriter := io.Pipe()
			multipartWriter := multipart.NewWriter(pipeWriter)
			wg.Add(1)
			go func() {
				defer multipartWriter.Close()
				defer wg.Done()
				// Create form data field "upload-file"
				multipartSrc, err := multipartWriter.CreateFormFile("upload-file", "./testdata/wink.png")
				if err != nil {
					t.Error(fmt.Errorf("%s: multipart writer create form file: %w", td.name, err))
				}
				srcFile, err := os.Open("./testdata/wink.png")
				if err != nil {
					t.Error(fmt.Errorf("%s: os open: %w", td.name, err))
				}
				defer srcFile.Close()
				imgSrc, _, err := image.Decode(srcFile)
				if err != nil {
					t.Error(fmt.Errorf("%s: image decode: %w", td.name, err))
				}
				if err := png.Encode(multipartSrc, imgSrc); err != nil {
					t.Error(fmt.Errorf("%s: png encode: %w", td.name, err))
				}
			}()
			r := httptest.NewRequest(http.MethodPost, "/", pipeReader)
			r.Header.Add("Content-Type", multipartWriter.FormDataContentType())
			toolkitTools := toolkit.Tools{
				AllowedFileTypes: td.allowedTypes,
			}
			// err used in a few places below
			uploadedFiles, err := toolkitTools.UploadFiles(r, "./testdata/uploads/", td.renameFile)
			if len(uploadedFiles) > 0 {
				t.Log(fmt.Sprintf("uploaded file: %s", uploadedFiles[0].NewFileName))
			}

			if err != nil && !td.errExpected {
				t.Error(fmt.Errorf("%s: upload files: %w", td.name, err))
			}

			if !td.errExpected {
				if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
					t.Error(fmt.Errorf("%s: expected file to exist: %w", td.name, err))
				}

				// Cleanup
				if err := os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); err != nil {
					t.Error(fmt.Errorf("%s: os remove: %w", td.name, err))
				}
			}

			if !td.errExpected && err != nil {
				t.Error(fmt.Errorf("%s: err expected but not received: %w", td.name, err))
			}

			wg.Wait()
		})
	}
}
