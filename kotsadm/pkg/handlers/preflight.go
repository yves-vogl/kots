package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/replicatedhq/kots/kotsadm/pkg/logger"
	"github.com/replicatedhq/kots/kotsadm/pkg/preflight"
	preflighttypes "github.com/replicatedhq/kots/kotsadm/pkg/preflight/types"
	"github.com/replicatedhq/kots/kotsadm/pkg/store"
	"github.com/replicatedhq/kots/pkg/kotsutil"
)

type GetPreflightResultResponse struct {
	PreflightResult preflighttypes.PreflightResult `json:"preflightResult"`
}

type GetPreflightCommandRequest struct {
	Origin string `json:"origin"`
}

type GetPreflightCommandResponse struct {
	Command []string `json:"command"`
}

func GetPreflightResult(w http.ResponseWriter, r *http.Request) {
	appSlug := mux.Vars(r)["appSlug"]
	sequence, err := strconv.ParseInt(mux.Vars(r)["sequence"], 10, 64)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(400)
		return
	}

	foundApp, err := store.GetStore().GetAppFromSlug(appSlug)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	result, err := store.GetStore().GetPreflightResults(foundApp.ID, sequence)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	response := GetPreflightResultResponse{
		PreflightResult: *result,
	}
	JSON(w, 200, response)
}

func GetLatestPreflightResult(w http.ResponseWriter, r *http.Request) {
	result, err := store.GetStore().GetLatestPreflightResults()
	if err != nil {
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	response := GetPreflightResultResponse{
		PreflightResult: *result,
	}
	JSON(w, 200, response)
}

func IgnorePreflightRBACErrors(w http.ResponseWriter, r *http.Request) {
	appSlug := mux.Vars(r)["appSlug"]
	sequence, err := strconv.Atoi(mux.Vars(r)["sequence"])
	if err != nil {
		logger.Error(err)
		w.WriteHeader(400)
		return
	}

	foundApp, err := store.GetStore().GetAppFromSlug(appSlug)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	if err := store.GetStore().SetIgnorePreflightPermissionErrors(foundApp.ID, int64(sequence)); err != nil {
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	archiveDir, err := store.GetStore().GetAppVersionArchive(foundApp.ID, int64(sequence))
	if err != nil {
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	go func() {
		defer os.RemoveAll(archiveDir)
		if err := preflight.Run(foundApp.ID, int64(sequence), foundApp.IsAirgap, archiveDir); err != nil {
			logger.Error(err)
			return
		}
	}()

	JSON(w, 200, struct{}{})
}

func StartPreflightChecks(w http.ResponseWriter, r *http.Request) {
	appSlug := mux.Vars(r)["appSlug"]
	sequence, err := strconv.Atoi(mux.Vars(r)["sequence"])
	if err != nil {
		logger.Error(err)
		w.WriteHeader(400)
		return
	}

	foundApp, err := store.GetStore().GetAppFromSlug(appSlug)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	if err := store.GetStore().ResetPreflightResults(foundApp.ID, int64(sequence)); err != nil {
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	archiveDir, err := store.GetStore().GetAppVersionArchive(foundApp.ID, int64(sequence))
	if err != nil {
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	go func() {
		defer os.RemoveAll(archiveDir)
		if err := preflight.Run(foundApp.ID, int64(sequence), foundApp.IsAirgap, archiveDir); err != nil {
			logger.Error(err)
			return
		}
	}()

	JSON(w, 200, struct{}{})
}

func GetPreflightCommand(w http.ResponseWriter, r *http.Request) {
	appSlug := mux.Vars(r)["appSlug"]
	sequence, err := strconv.ParseInt(mux.Vars(r)["sequence"], 10, 64)
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to parse sequence"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	inCluster := r.URL.Query().Get("inCluster") == "true"

	getPreflightCommandRequest := GetPreflightCommandRequest{}
	if err := json.NewDecoder(r.Body).Decode(&getPreflightCommandRequest); err != nil {
		logger.Error(errors.Wrap(err, "failed to decode request body"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	foundApp, err := store.GetStore().GetAppFromSlug(appSlug)
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to get app"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	archivePath, err := store.GetStore().GetAppVersionArchive(foundApp.ID, sequence)
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to get app archive"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(archivePath)

	kotsKinds, err := kotsutil.LoadKotsKindsFromPath(archivePath)
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to load kots kinds"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = preflight.CreateRenderedSpec(foundApp.ID, sequence, getPreflightCommandRequest.Origin, inCluster, kotsKinds.Preflight)
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to render preflight spec"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := GetPreflightCommandResponse{
		Command: preflight.GetPreflightCommand(foundApp.Slug),
	}

	JSON(w, http.StatusOK, response)
}

// PostPreflightStatus route is UNAUTHENTICATED
// This request comes from the `kubectl preflight` command.
func PostPreflightStatus(w http.ResponseWriter, r *http.Request) {
	appSlug := mux.Vars(r)["appSlug"]
	sequence, err := strconv.ParseInt(mux.Vars(r)["sequence"], 10, 64)
	if err != nil {
		err = errors.Wrap(err, "failed to parse sequence")
		logger.Error(err)
		w.WriteHeader(400)
		return
	}

	foundApp, err := store.GetStore().GetAppFromSlug(appSlug)
	if err != nil {
		err = errors.Wrap(err, "failed to get app from slug")
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err = errors.Wrap(err, "failed to read request body")
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	if err := store.GetStore().SetPreflightResults(foundApp.ID, sequence, b); err != nil {
		err = errors.Wrap(err, "failed to set preflight results")
		logger.Error(err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
}
