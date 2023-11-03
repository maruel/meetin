// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HomeAssistant struct {
	url    string
	apiKey string
}

func NewHomeAssistant(host, apiKey string) (*HomeAssistant, error) {
	// https://developers.home-assistant.io/docs/api/rest/
	// https://www.home-assistant.io/integrations/light/
	h := &HomeAssistant{url: "http://" + host + "/api/", apiKey: apiKey}
	// Do a request to confirm it works.
	state := map[string]interface{}{}
	if err := h.getJSON("config", &state); err != nil {
		return nil, err
	}
	//logJSON("Home", state)
	return h, nil
}

// GetState returns the state for a fully qualified entity name, e.g.
// "light.foobar".
func (h *HomeAssistant) GetState(entity string) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	if err := h.getJSON("states/"+entity, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (h *HomeAssistant) CallService(action string, in interface{}) error {
	m := []interface{}{}
	err := h.postJSON("services/"+action, in, &m)
	if len(m) != 0 {
		// TODO(maruel): Probably an error all the time.
		logJSON("services/"+action, m)
	}
	return err
}

// getJSON returns the state for a fully qualified entity name, e.g.
// "light.foobar".
func (h *HomeAssistant) getJSON(rsc string, out interface{}) error {
	req, err := http.NewRequest("GET", h.url+rsc, nil)
	if err != nil {
		return fmt.Errorf("get %s: %w", rsc, err)
	}
	return doJSON(req, h.apiKey, out)
}

func (h *HomeAssistant) postJSON(rsc string, in, out interface{}) error {
	raw, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("post %s: %w", rsc, err)
	}
	req, err := http.NewRequest("POST", h.url+rsc, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("post %s: %w", rsc, err)
	}
	return doJSON(req, h.apiKey, out)
}

func doJSON(req *http.Request, apiKey string, out interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("do reading: %w", err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("do json %s with %q: %w", req.URL, b, err)
	}
	return nil
}
