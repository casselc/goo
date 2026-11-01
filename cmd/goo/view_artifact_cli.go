package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/casselc/goo/workbench"
)

const viewImportTimeout = 15 * time.Second

type viewArtifactCLIReceipt struct {
	Version   string                     `json:"version"`
	Namespace string                     `json:"namespace"`
	Name      string                     `json:"name"`
	Revision  string                     `json:"revision"`
	View      *workbench.ExplorationView `json:"view,omitempty"`
	ID        string                     `json:"id,omitempty"`
}

func viewArtifactCLI(command string, args []string, out io.Writer) error {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	file := fs.String("file", "", "portable view YAML file (at most 64 KiB)")
	var destination string
	if command == "import-view" {
		fs.StringVar(&destination, "url", "", "explicit destination UI base URL; appends /api/exploration-views")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *file == "" || fs.NArg() != 0 || command == "import-view" && destination == "" {
		required := "-file"
		if command == "import-view" {
			required += " and -url"
		}
		return fmt.Errorf("%s requires %s and no positional arguments", command, required)
	}
	// Complete offline parsing before even constructing a network request.
	artifact, err := readViewArtifactFile(*file)
	if err != nil {
		return fmt.Errorf("%s: %w", *file, err)
	}
	receipt := viewArtifactCLIReceipt{Version: "goo.view-validation/1", Namespace: artifact.Artifact.Namespace, Name: artifact.Artifact.Name, Revision: artifact.Artifact.Revision, View: &artifact.Artifact.View}
	if command == "import-view" {
		endpoint, err := viewArtifactDestination(destination)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), viewImportTimeout)
		defer cancel()
		client := &http.Client{Timeout: viewImportTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		id, err := submitViewArtifact(ctx, client, endpoint, artifact.Artifact.View)
		if err != nil {
			return fmt.Errorf("import-view: %w; destination may have saved the view; no retry was attempted", err)
		}
		receipt.Version = "goo.view-import/1"
		receipt.View = nil
		receipt.ID = id
	}
	if err := json.NewEncoder(out).Encode(receipt); err != nil {
		if receipt.ID != "" {
			return fmt.Errorf("view saved as %s, but writing receipt failed: %w", receipt.ID, err)
		}
		return err
	}
	return nil
}

func readViewArtifactFile(path string) (*workbench.ViewArtifactResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("require a regular YAML file")
	}
	raw, err := io.ReadAll(io.LimitReader(file, workbench.ViewArtifactBytes+1))
	if err != nil {
		return nil, err
	}
	return workbench.DecodeViewArtifact(string(raw))
}

func viewArtifactDestination(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.RawFragment != "" || u.Opaque != "" {
		return "", errors.New("-url requires an absolute http(s) UI base URL without credentials, query or fragment")
	}
	if strings.Contains(base, "#") {
		return "", errors.New("-url must omit fragments")
	}
	escaped := strings.TrimRight(u.EscapedPath(), "/") + "/api/exploration-views"
	u.Path, err = url.PathUnescape(escaped)
	if err != nil {
		return "", errors.New("-url has invalid escaped base path")
	}
	u.RawPath = escaped
	return u.String(), nil
}

func submitViewArtifact(ctx context.Context, client *http.Client, endpoint string, view workbench.ExplorationView) (string, error) {
	raw, err := json.Marshal(view)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	// Do not provide a replayable body: even transport-level POST retries are
	// outside this explicit one-submission command.
	req.GetBody = nil
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("destination request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("destination returned HTTP %d (expected 201)", response.StatusCode)
	}
	media, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return "", errors.New("destination returned a non-JSON receipt")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, workbench.MaxBytes+1))
	if err != nil {
		return "", fmt.Errorf("reading destination receipt: %w", err)
	}
	if len(body) > workbench.MaxBytes {
		return "", errors.New("destination receipt exceeds 128 KiB")
	}
	var saved workbench.ExplorationView
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&saved); err != nil {
		return "", fmt.Errorf("invalid destination receipt: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return "", errors.New("destination receipt requires one JSON object")
	}
	id := saved.ID
	saved.ID = ""
	if !workbench.ID.MatchString(id) {
		return "", errors.New("destination receipt has invalid saved view ID")
	}
	if err = saved.Validate(); err != nil {
		return "", fmt.Errorf("invalid saved view: %w", err)
	}
	canonical, err := json.Marshal(saved)
	if err != nil {
		return "", err
	}
	if !bytes.Equal(canonical, raw) {
		return "", errors.New("destination receipt changed the submitted fixed view")
	}
	// Store identity is SHA-256 of the ID-free JSON view, not the descriptive
	// namespace/name/revision labels on the portable file.
	expectedID, err := workbench.ViewContentID(view)
	if err != nil {
		return "", err
	}
	if id != expectedID {
		return "", errors.New("destination receipt ID does not match submitted view content")
	}
	return id, nil
}
