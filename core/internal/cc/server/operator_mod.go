package server

import (
	"encoding/json"
	"net/http"

	"github.com/jm33-m0/emp3r0r/core/internal/cc/modules"
	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
)

func handleSetActiveModule(wrt http.ResponseWriter, req *http.Request) {
	// Decode JSON request body
	operation, err := DecodeJSONBody[def.Operation](wrt, req)
	if err != nil {
		return
	}
	if !operation.IsOptionSet("module_name") {
		http.Error(wrt, "No module selected", http.StatusBadRequest)
		return
	}

	// Set active module
	modules.SetActiveModule(*operation.ModuleName)
	wrt.WriteHeader(http.StatusOK)
}

func handleModuleRun(wrt http.ResponseWriter, req *http.Request) {
	// Decode JSON request body
	operation, err := DecodeJSONBody[def.Operation](wrt, req)
	if err != nil {
		return
	}

	// Get active module
	if !operation.IsOptionSet("module_name") {
		http.Error(wrt, "No module selected", http.StatusBadRequest)
		return
	}

	// Send command to module
	modules.ModuleRun()
	wrt.WriteHeader(http.StatusOK)
}

func handleModuleSetOption(wrt http.ResponseWriter, req *http.Request) {
	// Decode JSON request body
	operation, err := DecodeJSONBody[def.Operation](wrt, req)
	if err != nil {
		return
	}
	if !operation.IsOptionSet("set_option") || !operation.IsOptionSet("set_value") {
		http.Error(wrt, "Option or value not set", http.StatusBadRequest)
		return
	}

	// Set module option
	live.SetOption(*operation.SetOption, *operation.SetValue)
	wrt.WriteHeader(http.StatusOK)
}

func handleListModules(wrt http.ResponseWriter, req *http.Request) {
	// List all modules
	mod_comment_map := make(map[string]string)
	for mod_name, mod := range def.Modules {
		mod_comment_map[mod_name] = mod.Comment
	}

	// Send response
	wrt.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wrt).Encode(mod_comment_map); err != nil {
		http.Error(wrt, err.Error(), http.StatusInternalServerError)
	}
}
