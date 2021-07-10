// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// meetin displays a meeting.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

func logJSON(context string, m interface{}) {
	b, _ := json.MarshalIndent(m, "  ", "  ")
	log.Printf("%s:\n%s", context, b)
}

func printCalendars(ctx context.Context, c *Calendar) error {
	s, err := c.GetCalendars(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Calendars:\n")
	for _, cc := range s.Items {
		fmt.Printf("- %s\n", cc.Id)
	}
	return nil
}

func mainImpl() error {
	q := flag.Bool("q", false, "quiet")
	p := flag.Int("p", 24, "number of pixels")
	hahost := flag.String("hahost", "homeassistant:8123", "Home Assistant host")
	cid := flag.String("cid", "primary", "Calendar ID")
	path := flag.String("path", ".", "Path that contains the config")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("unexpected arguments")
	}
	if *q {
		log.SetOutput(ioutil.Discard)
	}
	if *p <= 0 || *p > 1000000000 {
		return errors.New("invalid number of pixels")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan os.Signal)
	go func() {
		<-ch
		cancel()
	}()
	signal.Notify(ch, os.Interrupt)

	if *path != "." {
		if err := os.Chdir(*path); err != nil {
			return err
		}
	}
	b, err := os.ReadFile("api.key")
	if err != nil {
		return err
	}
	ha, err := NewHomeAssistant(*hahost, strings.TrimSpace(string(b)))
	if err != nil {
		return err
	}

	// Initialize calendar API client.
	// Inspired by https://developers.google.com/calendar/api/quickstart/go
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	c, err := NewCalendar(ctx, "credentials.json", "token.json")
	if err != nil {
		return err
	}
	if err = printCalendars(ctx, c); err != nil {
		return err
	}

	return runLoop(ctx, ha, c, *cid)
}

func runLoop(ctx context.Context, ha *HomeAssistant, c *Calendar, cid string) error {
	//state, err := ha.GetState("light.meetin_ring")
	//if err != nil {
	//	return err
	//}
	//logJSON("Light", state)
	dataOff := map[string]interface{}{
		"entity_id": "light.meetin_ring",
	}
	data := map[string]interface{}{
		"entity_id":  "light.meetin_ring",
		"brightness": 128,
	}
	if err := ha.CallService("light/turn_off", dataOff); err != nil {
		return err
	}
	for ctx.Err() == nil {
		// Round time to the next 29min55s
		now := time.Now()
		// The idea is that querying takes generally less than 10 seconds. But we
		// want to query as much at last minute as possible in case a last minute
		// meeting was added.
		nextRounded := now.Round(30 * time.Minute)
		if nextRounded.Before(now) {
			nextRounded = nextRounded.Add(30 * time.Minute)
		}
		nextEarly := nextRounded.Add(-10 * time.Second)
		if d := nextEarly.Sub(now); d > 0 {
			log.Printf("Sleeping for %s", d.Round(time.Second))
			select {
			case <-ctx.Done():
				break
			case <-time.After(d):
			}
		}
		if ctx.Err() != nil {
			break
		}

		// Only take events that start on the time sharp. We want to limit the
		// number of RPCs per day to not hit any quota.
		events, err := c.GetEvents(ctx, nextRounded, time.Second, cid)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			//log.Println("  No upcoming events found.")
			continue
		}

		// Do one last sleep to align.
		now = time.Now()
		if d := nextRounded.Sub(now); d > 0 {
			select {
			case <-ctx.Done():
				break
			case <-time.After(d):
			}
		}
		if ctx.Err() != nil {
			break
		}

		var max time.Duration
		var end time.Time
		log.Println("Upcoming events:")
		for _, item := range events {
			d := item.end.Sub(item.start)
			log.Printf("- %s for %s\n", item.start.Format("2006-01-02T15:04:05"), d)
			// Events for more than 1h are not supported.
			if d > max && d <= time.Hour {
				max = d
				end = item.end
			}
		}
		if max == 0 {
			// Just got a long event, skip. Hack to wait past the :30.
			time.Sleep(11 * time.Second)
			continue
		}
		l := "30m"
		if max > 30*time.Minute {
			l = "60m"
		}
		// If the light was already on this effect, nothing will change. So turn it
		// blue then on the effect.
		log.Printf("Turning light for %s", l)
		if err := ha.CallService("light/turn_off", dataOff); err != nil {
			return err
		}
		data["effect"] = l
		if err := ha.CallService("light/turn_on", data); err != nil {
			return err
		}

		// Sleep until the meeting is done.
		now = time.Now()
		if d := end.Sub(now) - 10*time.Second; d > 0 {
			select {
			case <-ctx.Done():
				break
			case <-time.After(d):
			}
		}
		if ctx.Err() != nil {
			break
		}
	}
	// When the context is canceled, turn the light blue.
	if err := ha.CallService("light/turn_off", dataOff); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "meetin: %s\n", err)
		os.Exit(1)
	}
}
