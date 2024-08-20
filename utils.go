package ttapi

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// H is a hashmap
type H map[string]any

// SGo stands for Safe Go or Shit Go depending how you feel about goroutine panic handling
// Basically just a wrapper around the built-in keyword "go" with crash recovery
func SGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Error("unexpected crash", r)
				debug.PrintStack()
			}
		}()
		fn()
	}()
}

func refStr(s string) *string {
	return &s
}

// Sha1 returns sha1 hex sum as a string
func Sha1(in []byte) string {
	h := sha1.New()
	h.Write(in)
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateToken generate a random 32 bytes hex token
func GenerateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func safeMapPath(m any, path string) (out any) {
	parts := strings.Split(path, ".")
	tmp := m
	for _, p := range parts {
		if newMap, ok := tmp.(map[string]any); ok {
			out = newMap[p]
			tmp = out
		} else {
			return nil
		}
	}
	return
}

func castStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func doParseInt(s string) (out int) {
	out, _ = strconv.Atoi(s)
	return
}

func truncStr(s string, maxLen int, suffix string) (out string) {
	if len(s) <= maxLen {
		return s
	}
	out = s[0:maxLen]
	if suffix != "" {
		out += suffix
	}
	return
}

func isValidStatus(status string) bool {
	return status == available ||
		status == unavailable ||
		status == away
}

func isValidLaptop(laptop string) bool {
	return laptop == androidLaptop ||
		laptop == chromeLaptop ||
		laptop == iphoneLaptop ||
		laptop == linuxLaptop ||
		laptop == macLaptop ||
		laptop == pcLaptop
}

// Ternary ...
func Ternary[T any](predicate bool, a, b T) T {
	if predicate {
		return a
	}
	return b
}

// Or return "a" if it is non-zero otherwise "b"
func Or[T comparable](a, b T) (zero T) {
	return Ternary(a != zero, a, b)
}
