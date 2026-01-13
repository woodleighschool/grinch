package sync

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/woodleighschool/grinch/internal/domain/errx"
)

const (
	contentTypeJSON     = "application/json"
	contentTypeProtobuf = "application/x-protobuf"
)

// decodeRequest decodes the request body into msg, handling optional compression and JSON/protobuf payloads.
func decodeRequest[T proto.Message](r *http.Request, msg T) error {
	body, err := readBody(r)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}

	if isProtobuf(r.Header.Get("Content-Type")) {
		return decodeProtobuf(body, msg)
	}
	return decodeJSON(body, msg)
}

// readBody reads the request body and applies Content-Encoding decompression when present.
func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()

	encoding := strings.ToLower(r.Header.Get("Content-Encoding"))
	reader, err := decompressReader(r.Body, encoding)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errx.WithStatus(fmt.Errorf("read body: %w", err), http.StatusBadRequest)
	}
	return data, nil
}

// decompressReader wraps body with a decompressor based on Content-Encoding.
func decompressReader(body io.Reader, encoding string) (io.Reader, error) {
	switch encoding {
	case "", "identity":
		return body, nil
	case "gzip":
		gr, err := gzip.NewReader(body)
		if err != nil {
			return nil, errx.WithStatus(fmt.Errorf("gzip: %w", err), http.StatusBadRequest)
		}
		return gr, nil
	case "deflate":
		zr, err := zlib.NewReader(body)
		if err != nil {
			return nil, errx.WithStatus(fmt.Errorf("deflate: %w", err), http.StatusBadRequest)
		}
		return zr, nil
	default:
		return nil, errx.WithStatus(fmt.Errorf("unsupported encoding: %s", encoding), http.StatusUnsupportedMediaType)
	}
}

func isProtobuf(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "protobuf")
}

func decodeProtobuf[T proto.Message](data []byte, msg T) error {
	if err := proto.Unmarshal(data, msg); err != nil {
		return errx.WithStatus(fmt.Errorf("protobuf decode: %w", err), http.StatusBadRequest)
	}
	return nil
}

func decodeJSON[T proto.Message](data []byte, msg T) error {
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal(data, msg); err != nil {
		return errx.WithStatus(fmt.Errorf("json decode: %w", err), http.StatusBadRequest)
	}
	return nil
}

// encodeResponse encodes msg using the request Content-Type.
func encodeResponse(w http.ResponseWriter, r *http.Request, msg proto.Message) error {
	if isProtobuf(r.Header.Get("Content-Type")) {
		return writeProtobuf(w, msg)
	}
	return writeJSON(w, msg)
}

func writeProtobuf(w http.ResponseWriter, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("protobuf encode: %w", err)
	}
	w.Header().Set("Content-Type", contentTypeProtobuf)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	return err
}

func writeJSON(w http.ResponseWriter, msg proto.Message) error {
	opts := protojson.MarshalOptions{
		EmitUnpopulated: false,
		UseEnumNumbers:  true,
	}
	data, err := opts.Marshal(msg)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	return err
}
