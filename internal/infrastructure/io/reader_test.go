package io

// test file for StreamReader

import (
	"testing"

	"github.com/Valery223/Dungeon/internal/domain"
)

func TestParseLineSimple(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    domain.IncomingEvent
		wantErr bool
	}{
		{
			name: "valid line with extra params",
			line: "[14:49:02] 1 9 extra1 extra2",
			want: domain.IncomingEvent{
				TimeSec:  14*3600 + 49*60 + 2,
				PlayerID: 1,
				ID:       9,
				Extra:    string("extra1 extra2"),
			},
			wantErr: false,
		},
		{
			name: "valid line without extra params",
			line: "[00:00:00] 42 2",
			want: domain.IncomingEvent{
				TimeSec:  0,
				PlayerID: 42,
				ID:       2,
				Extra:    string(""),
			},
			wantErr: false,
		},
		{
			name:    "invalid line with missing fields",
			line:    "[14:49:02] 1",
			want:    domain.IncomingEvent{},
			wantErr: true,
		},
		{
			name:    "invalid time format",
			line:    "14:49:02 1 100",
			want:    domain.IncomingEvent{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parseLineSimple(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLineSimple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if event != tt.want {
				t.Errorf("parseLineSimple() = %v, want %v", event, tt.want)
			}
		})
	}
}
