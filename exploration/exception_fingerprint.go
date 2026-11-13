package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	ExceptionFingerprintMaxTypeBytes  = 256
	ExceptionFingerprintMaxStackBytes = 8192
	ExceptionFingerprintMaxEvents     = 32
	ExceptionFingerprintMaxEventBytes = 65536
	exceptionFingerprintDomain        = "goo.exception-fingerprint/1\n"
)

type ExceptionInputs struct {
	Type       string `json:"type"`
	Stacktrace string `json:"stacktrace"`
}
type ExceptionFingerprint struct {
	Fingerprint string           `json:"fingerprint"`
	State       string           `json:"state"`
	Reason      string           `json:"reason"`
	Inputs      *ExceptionInputs `json:"inputs"`
}

var exceptionReasons = map[string]bool{"none": true, "both-missing": true, "missing-type": true, "missing-stack": true, "both-empty": true, "empty-type": true, "empty-stack": true, "multiple": true, "oversize": true, "invalid-utf8": true}

func unclassifiedException(reason string) ExceptionFingerprint {
	return ExceptionFingerprint{Fingerprint: "ef1:u:" + reason, State: "unclassified", Reason: reason}
}
func classifiedException(typ, stack string) ExceptionFingerprint {
	input := exceptionFingerprintDomain + strconv.Itoa(len(typ)) + ":" + typ + strconv.Itoa(len(stack)) + ":" + stack
	digest := sha256.Sum256([]byte(input))
	return ExceptionFingerprint{Fingerprint: "ef1:c:" + hex.EncodeToString(digest[:]), State: "classified", Inputs: &ExceptionInputs{typ, stack}}
}
func ValidateExceptionFingerprint(value string) error {
	if strings.HasPrefix(value, "ef1:u:") && exceptionReasons[strings.TrimPrefix(value, "ef1:u:")] {
		return nil
	}
	if len(value) == 70 && strings.HasPrefix(value, "ef1:c:") {
		raw, e := hex.DecodeString(value[6:])
		if e == nil && len(raw) == 32 && strings.ToLower(value) == value {
			return nil
		}
	}
	return fmt.Errorf("invalid exception fingerprint v1")
}
func exceptionPairReason(attrs map[string]string) string {
	typ, hasType := attrs["exception.type"]
	stack, hasStack := attrs["exception.stacktrace"]
	switch {
	case len(typ) > ExceptionFingerprintMaxTypeBytes || len(stack) > ExceptionFingerprintMaxStackBytes:
		return "oversize"
	case !utf8.ValidString(typ) || !utf8.ValidString(stack):
		return "invalid-utf8"
	case !hasType && !hasStack:
		return "both-missing"
	case !hasType:
		return "missing-type"
	case !hasStack:
		return "missing-stack"
	case typ == "" && stack == "":
		return "both-empty"
	case typ == "":
		return "empty-type"
	case stack == "":
		return "empty-stack"
	default:
		return ""
	}
}
func ReferenceLogExceptionFingerprint(attrs map[string]string) ExceptionFingerprint {
	if reason := exceptionPairReason(attrs); reason != "" {
		return unclassifiedException(reason)
	}
	return classifiedException(attrs["exception.type"], attrs["exception.stacktrace"])
}

// One input record contributes once. Exception events are never exploded into
// occurrence rows, and a type and stack from separate events are never paired.
func ReferenceSpanExceptionFingerprint(names []string, attrs []map[string]string) (ExceptionFingerprint, error) {
	if len(names) != len(attrs) {
		return ExceptionFingerprint{}, fmt.Errorf("exception event names and attributes must align")
	}
	type tuple struct {
		HasType  bool
		Type     string
		HasStack bool
		Stack    string
	}
	tuples := map[tuple]bool{}
	events, total := 0, 0
	oversize, invalid := false, false
	for i, name := range names {
		if name != "exception" {
			continue
		}
		events++
		if events > ExceptionFingerprintMaxEvents {
			return unclassifiedException("oversize"), nil
		}
		typ, ht := attrs[i]["exception.type"]
		stack, hs := attrs[i]["exception.stacktrace"]
		total += len(typ) + len(stack)
		oversize = oversize || len(typ) > ExceptionFingerprintMaxTypeBytes || len(stack) > ExceptionFingerprintMaxStackBytes
		if oversize || total > ExceptionFingerprintMaxEventBytes {
			return unclassifiedException("oversize"), nil
		}
		invalid = invalid || !utf8.ValidString(typ) || !utf8.ValidString(stack)
		tuples[tuple{ht, typ, hs, stack}] = true
	}
	if events == 0 {
		return unclassifiedException("none"), nil
	}
	if events > ExceptionFingerprintMaxEvents || total > ExceptionFingerprintMaxEventBytes || oversize {
		return unclassifiedException("oversize"), nil
	}
	if invalid {
		return unclassifiedException("invalid-utf8"), nil
	}
	if len(tuples) > 1 {
		return unclassifiedException("multiple"), nil
	}
	for v := range tuples {
		a := map[string]string{}
		if v.HasType {
			a["exception.type"] = v.Type
		}
		if v.HasStack {
			a["exception.stacktrace"] = v.Stack
		}
		return ReferenceLogExceptionFingerprint(a), nil
	}
	panic("nonempty exception tuple set")
}

// Each tuple is presence/type/presence/stack/oversize/invalid/raw byte size.
// Values fed to SHA256 are bounded even under eager ClickHouse branch evaluation.
func exceptionTupleSQL(attrs string) string {
	t := attrs + "['exception.type']"
	s := attrs + "['exception.stacktrace']"
	ct := "if(length(" + t + ") <= 256, " + t + ", '')"
	cs := "if(length(" + s + ") <= 8192, " + s + ", '')"
	return "tuple(mapContains(" + attrs + ", 'exception.type'), " + ct + ", mapContains(" + attrs + ", 'exception.stacktrace'), " + cs + ", length(" + t + ") > 256 OR length(" + s + ") > 8192, NOT isValidUTF8(" + ct + ") OR NOT isValidUTF8(" + cs + "), toUInt64(length(" + t + ")) + toUInt64(length(" + s + ")))"
}
func exceptionPairReasonSQL(p string) string {
	return "multiIf(NOT " + p + ".1 AND NOT " + p + ".3, 'both-missing', NOT " + p + ".1, 'missing-type', NOT " + p + ".3, 'missing-stack', " + p + ".2 = '' AND " + p + ".4 = '', 'both-empty', " + p + ".2 = '', 'empty-type', " + p + ".4 = '', 'empty-stack', '')"
}
func exceptionProjectionSQL(signal string) string {
	if signal == "logs" {
		tuple := exceptionTupleSQL("LogAttributes")
		return "arrayElement(arrayMap(p -> tuple(multiIf(p.5, 'oversize', p.6, 'invalid-utf8', " + exceptionPairReasonSQL("p") + "), p.2, p.4), [" + tuple + "]), 1)"
	}
	matching := "arrayFilter(e -> e.1 = 'exception', arrayZip(`Events.Name`, `Events.Attributes`))"
	prepared := "arrayMap(e -> " + exceptionTupleSQL("e.2") + ", arraySlice(" + matching + ", 1, 33))"
	reason := "multiIf(length(es) = 0, 'none', length(es) > 32 OR arrayExists(p -> p.5, es) OR arraySum(arrayMap(p -> p.7, es)) > 65536, 'oversize', arrayExists(p -> p.6, es), 'invalid-utf8', length(arrayDistinct(es)) > 1, 'multiple', " + exceptionPairReasonSQL("arrayElement(es, 1)") + ")"
	return "arrayElement(arrayMap(es -> tuple(" + reason + ", arrayElement(es, 1).2, arrayElement(es, 1).4), [" + prepared + "]), 1)"
}
func exceptionFingerprintFromProjectionSQL(projection string) string {
	// Explicit lengths frame the exact bytes; punctuation, newlines and Unicode
	// remain literal inputs. This domain intentionally does not include the signal.
	hash := "lower(hex(SHA256(concat('goo.exception-fingerprint/1\\n', toString(length(p.2)), ':', p.2, toString(length(p.3)), ':', p.3))))"
	return "arrayElement(arrayMap(p -> if(p.1 = '', concat('ef1:c:', " + hash + "), concat('ef1:u:', p.1)), [" + projection + "]), 1)"
}
func exceptionFingerprintSQL(signal string) string {
	return exceptionFingerprintFromProjectionSQL(exceptionProjectionSQL(signal))
}
