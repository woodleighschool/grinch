package apihttp //nolint:dupl // structurally similar to users.go by design

import (
	"net/http"
)

func (handler *Server) ListExecutables(
	writer http.ResponseWriter,
	request *http.Request,
	params ListExecutablesParams,
) {
	listOptions, err := parseListOptions(
		params.Limit,
		params.Offset,
		params.Search,
		params.Sort,
		params.Order,
		params.Ids,
	)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	items, total, err := handler.store.ListExecutables(request.Context(), listOptions)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, ExecutableListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) GetExecutable(writer http.ResponseWriter, request *http.Request, id Id) {
	executable, err := handler.store.GetExecutable(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "executable not found"})
		return
	}

	writeJSON(writer, http.StatusOK, executable)
}
