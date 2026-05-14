package handler

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"converter/internal/converter"
)

const (
	maxJSONBodySize = 2 << 20
	maxUploadSize   = 32 << 20
	uploadFormField = "file"
)

type convertRequest struct {
	Input  string `json:"input"`
	Target string `json:"target"` // "cpa" | "sub2" | "auto"
}

type convertResponse struct {
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Detected string `json:"detected,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func convertBytes(input []byte, target string) ([]byte, string, error) {
	detected := converter.DetectFormat(input)
	if target == "auto" || target == "" {
		switch detected {
		case "cpa":
			target = "sub2"
		case "sub2":
			target = "cpa"
		default:
			return nil, detected, fmt.Errorf("cannot detect source format automatically")
		}
	}

	switch target {
	case "sub2":
		out, err := converter.CPAToSub2(input)
		return out, detected, err
	case "cpa":
		out, err := converter.Sub2ToCPA(input)
		return out, detected, err
	default:
		return nil, detected, fmt.Errorf("unknown target format: %s", target)
	}
}

// ConvertHandler handles POST /api/convert
func ConvertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, convertResponse{Error: "method not allowed"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxJSONBodySize))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "failed to read request body"})
		return
	}

	var req convertRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "invalid JSON payload"})
		return
	}

	if strings.TrimSpace(req.Input) == "" {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "input cannot be empty"})
		return
	}

	out, detected, err := convertBytes([]byte(req.Input), req.Target)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, convertResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, convertResponse{
		Output:   string(out),
		Detected: detected,
	})
}

// DetectHandler handles POST /api/detect
func DetectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, convertResponse{Error: "method not allowed"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxJSONBodySize))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "failed to read request body"})
		return
	}

	var req struct {
		Input string `json:"input"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "invalid JSON payload"})
		return
	}

	detected := converter.DetectFormat([]byte(req.Input))
	writeJSON(w, http.StatusOK, map[string]string{"format": detected})
}

// ConvertFileHandler handles POST /api/convert-file
func ConvertFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, convertResponse{Error: "method not allowed"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "failed to parse upload, make sure the file is under 32MB"})
		return
	}

	file, header, err := r.FormFile(uploadFormField)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "please upload a .json or .zip file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "failed to read uploaded file"})
		return
	}

	target := r.FormValue("target")
	name := header.Filename
	ext := strings.ToLower(path.Ext(name))

	switch ext {
	case ".json":
		serveConvertedJSON(w, name, data, target)
	case ".zip":
		serveConvertedZIP(w, data, target)
	default:
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "only .json and .zip files are supported"})
	}
}

func serveConvertedJSON(w http.ResponseWriter, filename string, data []byte, target string) {
	out, _, err := convertBytes(data, target)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, convertResponse{Error: err.Error()})
		return
	}

	downloadName := strings.TrimSuffix(path.Base(filename), path.Ext(filename)) + "_converted.json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", downloadName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func serveConvertedZIP(w http.ResponseWriter, data []byte, target string) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "invalid zip archive"})
		return
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	convertedCount := 0

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if strings.ToLower(path.Ext(file.Name)) != ".json" {
			continue
		}

		src, err := file.Open()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, convertResponse{Error: "failed to read file inside zip: " + file.Name})
			_ = zipWriter.Close()
			return
		}

		content, readErr := io.ReadAll(src)
		_ = src.Close()
		if readErr != nil {
			writeJSON(w, http.StatusBadRequest, convertResponse{Error: "failed to read file inside zip: " + file.Name})
			_ = zipWriter.Close()
			return
		}

		out, _, err := convertBytes(content, target)
		if err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, convertResponse{Error: fmt.Sprintf("%s conversion failed: %v", file.Name, err)})
			_ = zipWriter.Close()
			return
		}

		entryName := strings.TrimSuffix(file.Name, path.Ext(file.Name)) + "_converted.json"
		dst, err := zipWriter.Create(entryName)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, convertResponse{Error: "failed to build output zip"})
			_ = zipWriter.Close()
			return
		}
		if _, err := dst.Write(out); err != nil {
			writeJSON(w, http.StatusInternalServerError, convertResponse{Error: "failed to write output zip"})
			_ = zipWriter.Close()
			return
		}
		convertedCount++
	}

	if convertedCount == 0 {
		writeJSON(w, http.StatusBadRequest, convertResponse{Error: "the zip file does not contain any convertible .json files"})
		_ = zipWriter.Close()
		return
	}

	if err := zipWriter.Close(); err != nil {
		writeJSON(w, http.StatusInternalServerError, convertResponse{Error: "failed to finalize output zip"})
		return
	}

	filename := fmt.Sprintf("converted_%s.zip", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
