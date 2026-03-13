package httpapi

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListExecutables(
	writer http.ResponseWriter,
	request *http.Request,
	params ListExecutablesParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	items, total, err := handler.admin.ListExecutables(request.Context(), domain.ExecutableListOptions{
		ListOptions: listOptions,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped := make([]ExecutableSummary, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapExecutableSummary(item))
	}

	writeJSON(writer, http.StatusOK, ExecutableListResponse{
		Rows:  mapped,
		Total: total,
	})
}

func (handler *Server) GetExecutable(writer http.ResponseWriter, request *http.Request, id Id) {
	executable, err := handler.admin.GetExecutable(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "executable not found"})
		return
	}

	writeJSON(writer, http.StatusOK, mapExecutable(executable))
}
