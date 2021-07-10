// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Calendar interfaces with the Google Calendar API with an OAuth2 service
// account and a specific user token.
type Calendar struct {
	srv *calendar.Service
}

// NewCalendar returns an initialized calendar client.
func NewCalendar(ctx context.Context, credentialsFile, tokenFile string) (*Calendar, error) {
	// https://developers.google.com/workspace/guides/create-credentials
	b, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	client, err := getGoogleAPIClient(ctx, tokenFile, config)
	if err != nil {
		return nil, err
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}
	return &Calendar{srv: srv}, nil
}

func (c *Calendar) GetCalendars(ctx context.Context) (*calendar.CalendarList, error) {
	return c.srv.CalendarList.List().Do()
}

// Event is a simplified calendar event.
type Event struct {
	summary string
	status  string
	start   time.Time
	end     time.Time
}

// GetEvents returns the next events starting between start and for the
// duration (exclusive).
func (c *Calendar) GetEvents(ctx context.Context, start time.Time, d time.Duration, id string) ([]Event, error) {
	t1 := start.Format(time.RFC3339)
	t2 := start.Add(d).Format(time.RFC3339)
	events, err := c.srv.Events.List(id).ShowDeleted(false).
		SingleEvents(true).TimeMin(t1).TimeMax(t2).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve next ten of the user's events: %w", err)
	}

	out := make([]Event, 0, len(events.Items))
	for _, item := range events.Items {
		if item.Status == "cancelled" {
			continue
		}
		start := item.Start.DateTime
		if start == "" {
			// Ignore all day event. They only have item.Start.Date set.
			continue
		}
		s, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			return out, err
		}
		e, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			return out, err
		}
		out = append(out, Event{
			summary: item.Summary,
			status:  item.Status,
			start:   s,
			end:     e,
		})
	}
	return out, nil
}

// getGoogleAPIClient retrieves a token, saves the token, then returns the generated
// client.
func getGoogleAPIClient(ctx context.Context, tf string, config *oauth2.Config) (*http.Client, error) {
	tok, err := getOauth2TokenFromFile(tf)
	if err != nil {
		if tok, err = getOauth2TokenFromWeb(config); err != nil {
			return nil, err
		}
		if err = saveToken(tf, tok); err != nil {
			return nil, err
		}
	}
	return config.Client(ctx, tok), nil
}

// getOauth2TokenFromWeb requests a token from the web, then returns the retrieved
// token.
func getOauth2TokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	u := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following URL in your browser then type the authorization code: \n%v\n", u)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

// getOauth2TokenFromFile retrieves a token from a local file.
func getOauth2TokenFromFile(tf string) (*oauth2.Token, error) {
	d, err := ioutil.ReadFile(tf)
	if err == nil {
		tok := &oauth2.Token{}
		if err = json.Unmarshal(d, tok); err == nil {
			return tok, nil
		}
	}
	return nil, err
}

// saveToken saves a token to a file path.
func saveToken(tf string, token *oauth2.Token) error {
	f, err := os.OpenFile(tf, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}
