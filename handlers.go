package main

import (
  "encoding/json"
  "fmt"
  "net/http"

  "github.com/docker/docker/daemon/logger"
)

// startLoggingRequest represents the request object we get on a call to //LogDriver.StartLogging
type startLoggingRequest struct {
  File string
  Info logger.Info
}

// stopLoggingRequest represents the request object we get on a call to //LogDriver.StopLogging
type stopLoggingRequest struct {
  File string
}

// capabilitiesResponse represents the response to a capabilities request
type capabilitiesResponse struct {
  Err string
  Cap logger.Capability
}

type response struct {
  Err string
}

func reportCaps() func(w http.ResponseWriter, r *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(&capabilitiesResponse{
      Cap: logger.Capability{},
    })
  }
}

func startLoggingHandler() func(w http.ResponseWriter, r *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    var req startLoggingRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
      http.Error(w, fmt.Sprintf("error decoding json startLoggingRequest: %v", err), http.StatusBadRequest)
      return
    }

    fmt.Printf("startLoggingRequest: %#v", req)

    respondOK(w)
  }
}

func stopLoggingHandler() func(w http.ResponseWriter, r *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    var req stopLoggingRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
      http.Error(w, fmt.Errorf("error decoding json stopLoggingRequest: %w", err).Error(), http.StatusBadRequest)
      return
    }

    fmt.Printf("stopLoggingRequest: %#v", req)

    respondOK(w)
  }
}

func respondOK(w http.ResponseWriter) {
  var res response

  json.NewEncoder(w).Encode(&res)
}
