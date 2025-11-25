package santa

import (
	"encoding/json"
	"net/http"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type wireFormat int

const (
	wireFormatUnknown wireFormat = iota
	wireFormatJSON
	wireFormatProtobuf
)

var (
	protoJSONUnmarshal = protojson.UnmarshalOptions{DiscardUnknown: true}
	protoJSONMarshal   = protojson.MarshalOptions{UseProtoNames: true}
)

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func respondProto(w http.ResponseWriter, r *http.Request, status int, payload proto.Message) {
	format := responseWireFormat(r)
	contentType := contentTypeForFormat(format)
	var encoded []byte
	var err error
	switch format {
	case wireFormatProtobuf:
		if payload != nil {
			encoded, err = proto.Marshal(payload)
		}
	default:
		if payload != nil {
			encoded, err = protoJSONMarshal.Marshal(payload)
		}
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	if len(encoded) > 0 {
		_, _ = w.Write(encoded)
		if format == wireFormatJSON && encoded[len(encoded)-1] != '\n' {
			_, _ = w.Write([]byte("\n"))
		}
	}
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func unmarshalProtoJSON(data []byte, msg proto.Message) error {
	if len(data) == 0 {
		return nil
	}
	return protoJSONUnmarshal.Unmarshal(data, msg)
}

func marshalProtoJSON(msg proto.Message) ([]byte, error) {
	return protoJSONMarshal.Marshal(msg)
}

func decodeWireMessage(format wireFormat, data []byte, msg proto.Message) error {
	if len(data) == 0 {
		return nil
	}
	switch format {
	case wireFormatProtobuf:
		return proto.Unmarshal(data, msg)
	default:
		return unmarshalProtoJSON(data, msg)
	}
}

func requestWireFormat(r *http.Request) wireFormat {
	if format := wireFormatFromContentType(r.Header.Get("Content-Type")); format != wireFormatUnknown {
		return format
	}
	return wireFormatJSON
}

func responseWireFormat(r *http.Request) wireFormat {
	if format := wireFormatFromAccept(r.Header.Get("Accept")); format != wireFormatUnknown {
		return format
	}
	return requestWireFormat(r)
}

func wireFormatFromContentType(value string) wireFormat {
	media := strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	switch media {
	case "application/x-protobuf", "application/octet-stream":
		return wireFormatProtobuf
	case "application/json", "application/*+json", "":
		return wireFormatJSON
	default:
		if strings.HasSuffix(media, "+json") {
			return wireFormatJSON
		}
	}
	return wireFormatUnknown
}

func wireFormatFromAccept(value string) wireFormat {
	for _, part := range strings.Split(value, ",") {
		media := strings.ToLower(strings.TrimSpace(strings.Split(part, ";")[0]))
		switch media {
		case "application/x-protobuf", "application/octet-stream":
			return wireFormatProtobuf
		case "application/json", "application/*+json", "*/*", "":
			return wireFormatJSON
		default:
			if strings.HasSuffix(media, "+json") {
				return wireFormatJSON
			}
		}
	}
	return wireFormatUnknown
}

func contentTypeForFormat(format wireFormat) string {
	switch format {
	case wireFormatProtobuf:
		return "application/x-protobuf"
	default:
		return "application/json"
	}
}
