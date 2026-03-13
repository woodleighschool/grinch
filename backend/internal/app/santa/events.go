package santa

import (
	"context"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
)

// HandleEventUpload ingests execution and file access events.
func (service *Service) HandleEventUpload(
	ctx context.Context,
	machineID uuid.UUID,
	request *syncv1.EventUploadRequest,
) (*syncv1.EventUploadResponse, error) {
	// TODO: persist audit_events?
	ingested, err := service.dataStore.IngestEvents(
		ctx,
		machineID,
		request.GetEvents(),
		request.GetFileAccessEvents(),
		service.eventAllowlist,
	)
	if err != nil {
		return nil, err
	}

	_ = ingested
	return syncv1.EventUploadResponse_builder{}.Build(), nil
}
