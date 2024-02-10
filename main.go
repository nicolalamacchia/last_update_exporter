package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func probeHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	target := params.Get("target")

	if target == "" {
		http.Error(w, "Parameter 'target' is missing", http.StatusBadRequest)
		return
	}

	mostRecentUpdate, err := getMostRecentFileModTime(target)
	if err != nil {
		http.Error(w, "Error reading target folder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the time passed since the most recent update.
	timePassed := time.Now().Sub(mostRecentUpdate).Seconds()

	// Create a new registry to avoid metric collisions on concurrent probes.
	reg := prometheus.NewRegistry()
	folderUpdateMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "folder_last_update_seconds",
			Help: "Time in seconds since the last update of the probed folder.",
		},
		[]string{"folder"},
	)
	reg.MustRegister(folderUpdateMetric)

	// Set and collect the metric.
	folderUpdateMetric.WithLabelValues(target).Set(timePassed)
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func getMostRecentFileModTime(dirPath string) (time.Time, error) {
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return time.Time{}, err // return zero time and the error if any
	}

	return fileInfo.ModTime(), nil
}

const port int = 9188

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	slog.Info("Server starting", "port", port)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probeHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
