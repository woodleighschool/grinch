package santa

import (
	"encoding/json"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

func respondProtoJSON(w http.ResponseWriter, status int, payload proto.Message) {
	if payload == nil {
		respondJSON(w, status, nil)
		return
	}
	encoded, err := protoJSONMarshal.Marshal(payload)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if len(encoded) > 0 {
		_, _ = w.Write(encoded)
		if encoded[len(encoded)-1] != '\n' {
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
