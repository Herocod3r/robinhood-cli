package datearg

import (
	"testing"
	"time"
)

func TestParse_Absolute(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	got, err := Parse("2025-06-15", now)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParse_YTD(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	got, err := Parse("ytd", now)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParse_Days(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	got, err := Parse("7d", now)
	if err != nil {
		t.Fatal(err)
	}
	want := now.UTC().Add(-7 * 24 * time.Hour)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParse_Weeks(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	got, err := Parse("2w", now)
	if err != nil {
		t.Fatal(err)
	}
	want := now.UTC().Add(-14 * 24 * time.Hour)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParse_Months(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	got, err := Parse("3mo", now)
	if err != nil {
		t.Fatal(err)
	}
	want := now.UTC().Add(-90 * 24 * time.Hour)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParse_Years(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	got, err := Parse("1y", now)
	if err != nil {
		t.Fatal(err)
	}
	want := now.UTC().Add(-365 * 24 * time.Hour)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParse_Empty(t *testing.T) {
	if _, err := Parse("", time.Now()); err == nil {
		t.Fatal("empty string should error")
	}
}

func TestParse_Garbage(t *testing.T) {
	if _, err := Parse("banana", time.Now()); err == nil {
		t.Fatal("garbage should error")
	}
}

func TestParse_BadNumber(t *testing.T) {
	if _, err := Parse("xd", time.Now()); err == nil {
		t.Fatal("bad number should error")
	}
}

func TestParse_Negative(t *testing.T) {
	if _, err := Parse("-1d", time.Now()); err == nil {
		t.Fatal("negative duration should error")
	}
}
