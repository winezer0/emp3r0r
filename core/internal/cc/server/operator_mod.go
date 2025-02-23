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
	if err := json.NewEncoder(wrt).Encode(live.ActiveModule); err != nil {
		http.Error(wrt, err.Error(), http.StatusInternalServerError)
	}
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

// handleListModules returns a list of def.ModuleConfig
func handleListModules(wrt http.ResponseWriter, _ *http.Request) {
	// Send response
	wrt.Header().Set("Content-Type", "application/json")

	modules := make([]*def.ModuleConfig, 0)
	for _, mod := range def.Modules {
		modules = append(modules, mod)
	}
	if err := json.NewEncoder(wrt).Encode(modules); err != nil {
		http.Error(wrt, err.Error(), http.StatusInternalServerError)
	}
}

// handleSearchModule searches for a module and return a list of def.ModuleConfig
func handleSearchModule(wrt http.ResponseWriter, req *http.Request) {
	// Decode JSON request body
	operation, err := DecodeJSONBody[def.Operation](wrt, req)
	if err != nil {
		return
	}
	if !operation.IsOptionSet("module_name") {
		http.Error(wrt, "No module selected", http.StatusBadRequest)
		return
	}

	// Search module
	results := modules.ModuleSearch(*operation.ModuleName)
	wrt.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wrt).Encode(results); err != nil {
		http.Error(wrt, err.Error(), http.StatusInternalServerError)
	}
}
