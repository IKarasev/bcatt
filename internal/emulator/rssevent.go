package emulator

import (
	"bytes"
	"fmt"
	"io"
)

type RssEvent struct {
	ID      []byte
	Data    []byte
	Event   []byte
	Retry   []byte
	Comment []byte
}

type RssChan chan RssEvent

func NewRssEvent() *RssEvent {
	return &RssEvent{
		ID:      make([]byte, 0),
		Data:    make([]byte, 0),
		Event:   make([]byte, 0),
		Retry:   make([]byte, 0),
		Comment: make([]byte, 0),
	}
}

func (ev *RssEvent) WithId(b []byte) *RssEvent {
	ev.ID = b
	return ev
}

func (ev *RssEvent) WithData(b []byte) *RssEvent {
	ev.Data = b
	return ev
}
func (ev *RssEvent) WithEvent(b []byte) *RssEvent {
	ev.Event = b
	return ev
}
func (ev *RssEvent) WithRetry(b []byte) *RssEvent {
	ev.Retry = b
	return ev
}
func (ev *RssEvent) WithComment(b []byte) *RssEvent {
	ev.Comment = b
	return ev
}

// MarshalTo marshals Event to given Writer
func (ev *RssEvent) MarshalTo(w io.Writer) error {
	// Marshalling part is taken from: https://github.com/r3labs/sse/blob/c6d5381ee3ca63828b321c16baa008fd6c0b4564/http.go#L16
	if len(ev.Data) == 0 && len(ev.Comment) == 0 {
		return nil
	}

	if len(ev.Data) > 0 {
		if _, err := fmt.Fprintf(w, "id: %s\n", ev.ID); err != nil {
			return err
		}

		sd := bytes.Split(ev.Data, []byte("\n"))
		for i := range sd {
			if _, err := fmt.Fprintf(w, "data: %s\n", sd[i]); err != nil {
				return err
			}
		}

		if len(ev.Event) > 0 {
			if _, err := fmt.Fprintf(w, "event: %s\n", ev.Event); err != nil {
				return err
			}
		}

		if len(ev.Retry) > 0 {
			if _, err := fmt.Fprintf(w, "retry: %s\n", ev.Retry); err != nil {
				return err
			}
		}
	}

	if len(ev.Comment) > 0 {
		if _, err := fmt.Fprintf(w, ": %s\n", ev.Comment); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}

	return nil
}
