package apihttp

import "net/http"

func (s *Server) AddGroupUser(w http.ResponseWriter, r *http.Request, id Id, userID UserIdPath) {
	if err := s.groups.AddUser(r.Context(), id, userID); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}

func (s *Server) RemoveGroupUser(w http.ResponseWriter, r *http.Request, id Id, userID UserIdPath) {
	if err := s.groups.RemoveUser(r.Context(), id, userID); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}

func (s *Server) AddGroupMachine(w http.ResponseWriter, r *http.Request, id Id, machineID MachineIdPath) {
	if err := s.groups.AddMachine(r.Context(), id, machineID); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}

func (s *Server) RemoveGroupMachine(w http.ResponseWriter, r *http.Request, id Id, machineID MachineIdPath) {
	if err := s.groups.RemoveMachine(r.Context(), id, machineID); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}
