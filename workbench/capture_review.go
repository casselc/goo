package workbench

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

const CaptureReviewExportMaxBytes = 4 << 20
const captureReviewOrigin = "local-user-unverified"

type CaptureReviewResolver func(exploration.CaptureRecordRef) (*exploration.CapturedRecord, error)
type CaptureReviewLabel struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}
type CaptureReviewTaskInput struct {
	Version int                  `json:"version"`
	Name    string               `json:"name"`
	Title   string               `json:"title"`
	Labels  []CaptureReviewLabel `json:"labels"`
}
type CaptureReviewTask struct {
	ID         string               `json:"id"`
	Version    int                  `json:"version"`
	Name       string               `json:"name"`
	Title      string               `json:"title"`
	Labels     []CaptureReviewLabel `json:"labels"`
	RecordedAt string               `json:"recordedAt"`
	Origin     string               `json:"origin"`
}
type CaptureReviewJudgment struct {
	Kind  string `json:"kind"`
	Value string `json:"value,omitempty"`
}
type CaptureReviewDecisionInput struct {
	TaskID   string                       `json:"taskID"`
	Target   exploration.CaptureRecordRef `json:"target"`
	Parents  []string                     `json:"parents"`
	Judgment CaptureReviewJudgment        `json:"judgment"`
	Review   string                       `json:"review"`
	State    string                       `json:"state"`
	Note     string                       `json:"note"`
}
type CaptureReviewEvidence struct {
	Reference        exploration.CaptureRecordRef `json:"reference"`
	Receipt          exploration.CaptureReceipt   `json:"receipt"`
	RecordDigest     string                       `json:"recordDigest"`
	SourceReadAt     string                       `json:"sourceReadAt"`
	Basis            string                       `json:"basis"`
	ObservationState string                       `json:"observationState"`
	MatchingRecords  string                       `json:"matchingRecords"`
	RetainedRecords  int                          `json:"retainedRecords"`
	RecordsLimited   bool                         `json:"recordsLimited"`
}
type CaptureReviewDecision struct {
	ID         string                       `json:"id"`
	Version    int                          `json:"version"`
	TaskID     string                       `json:"taskID"`
	Target     exploration.CaptureRecordRef `json:"target"`
	Parents    []string                     `json:"parents"`
	Judgment   CaptureReviewJudgment        `json:"judgment"`
	Review     string                       `json:"review"`
	State      string                       `json:"state"`
	Note       string                       `json:"note"`
	Evidence   CaptureReviewEvidence        `json:"evidence"`
	RecordedAt string                       `json:"recordedAt"`
	Origin     string                       `json:"origin"`
}
type CaptureReviewManifestInput struct {
	TaskID     string   `json:"taskID"`
	Selected   []string `json:"selected"`
	KnownHeads []string `json:"knownHeads"`
	Historical bool     `json:"historical"`
}
type CaptureReviewManifestItem struct {
	DecisionID string                `json:"decisionID"`
	Label      string                `json:"label"`
	Evidence   CaptureReviewEvidence `json:"evidence"`
}
type CaptureReviewManifest struct {
	ID         string                      `json:"id"`
	Version    int                         `json:"version"`
	TaskID     string                      `json:"taskID"`
	Selected   []string                    `json:"selected"`
	KnownHeads []string                    `json:"knownHeads"`
	Historical bool                        `json:"historical"`
	Revisions  []string                    `json:"revisions"`
	Items      []CaptureReviewManifestItem `json:"items"`
	RecordedAt string                      `json:"recordedAt"`
	Origin     string                      `json:"origin"`
	Mode       string                      `json:"mode"`
}

func (in CaptureReviewTaskInput) Validate() error {
	if in.Version != 1 || !validTag(in.Name) || !text(in.Title, 120, true, false) || len(in.Labels) < 1 || len(in.Labels) > 16 {
		return errors.New("choose one versioned task with 1..16 labels")
	}
	seen := map[string]bool{}
	for _, label := range in.Labels {
		if !validTag(label.Key) || !text(label.Title, 120, true, false) || seen[label.Key] {
			return errors.New("invalid or duplicate task label")
		}
		seen[label.Key] = true
	}
	return nil
}
func reviewIDs(ids []string, max int) error {
	if len(ids) > max {
		return errors.New("review reference limit exceeded")
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if !ID.MatchString(id) || seen[id] {
			return errors.New("invalid or duplicate review reference")
		}
		seen[id] = true
	}
	return nil
}
func reviewTarget(ref exploration.CaptureRecordRef) error {
	if !ID.MatchString(ref.CaptureID) || ref.Index < 0 || ref.Index >= exploration.CaptureRowLimit {
		return errors.New("invalid captured occurrence")
	}
	return nil
}
func (in CaptureReviewDecisionInput) Validate() error {
	if !ID.MatchString(in.TaskID) || reviewTarget(in.Target) != nil || reviewIDs(in.Parents, 8) != nil || !text(in.Note, 2048, false, true) {
		return errors.New("invalid review decision input")
	}
	if !slices.Contains([]string{"unreviewed", "reviewed", "disputed"}, in.Review) || !slices.Contains([]string{"active", "withdrawn"}, in.State) || in.State == "withdrawn" && len(in.Parents) == 0 {
		return errors.New("choose explicit review and withdrawal states")
	}
	switch in.Judgment.Kind {
	case "label":
		if !validTag(in.Judgment.Value) {
			return errors.New("choose an explicit vocabulary label")
		}
	case "uncertain", "abstain":
		if in.Judgment.Value != "" {
			return errors.New("nonlabel judgments cannot carry a label")
		}
	default:
		return errors.New("unsupported judgment")
	}
	return nil
}
func (in CaptureReviewManifestInput) Validate() error {
	if !ID.MatchString(in.TaskID) || !in.Historical || len(in.Selected) < 1 || len(in.Selected) > 32 || reviewIDs(in.Selected, 32) != nil || reviewIDs(in.KnownHeads, 64) != nil {
		return errors.New("choose 1..32 pinned historical decisions and bounded known heads")
	}
	return nil
}
func reviewClock(at, origin string) bool { return validTime(at) && origin == captureReviewOrigin }
func reviewIntegrity(id string, body any) error {
	want, err := digest(body)
	if err != nil {
		return err
	}
	if !ID.MatchString(id) || want != id {
		return errors.New("capture review content integrity mismatch")
	}
	return nil
}
func (s Store) readReview(kind, id string, out any) error {
	var raw json.RawMessage
	if err := s.read(kind, id, &raw); err != nil {
		return err
	}
	if !identity.WellFormedJSON(raw) {
		return errors.New("invalid review Unicode")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	return d.Decode(out)
}
func (s Store) CreateCaptureReviewTask(in CaptureReviewTaskInput) (*CaptureReviewTask, error) {
	return s.createCaptureReviewTaskAt(in, time.Now().UTC().Format(time.RFC3339Nano))
}
func (s Store) createCaptureReviewTaskAt(in CaptureReviewTaskInput, at string) (*CaptureReviewTask, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	v := CaptureReviewTask{Version: in.Version, Name: in.Name, Title: in.Title, Labels: slices.Clone(in.Labels), RecordedAt: at, Origin: captureReviewOrigin}
	id, err := digest(v)
	if err != nil {
		return nil, err
	}
	v.ID = id
	if err = s.publish("capture-review-tasks", id, v); err != nil {
		return nil, err
	}
	return s.LoadCaptureReviewTask(id)
}
func (s Store) LoadCaptureReviewTask(id string) (*CaptureReviewTask, error) {
	var v CaptureReviewTask
	if err := s.readReview("capture-review-tasks", id, &v); err != nil {
		return nil, err
	}
	actual := v.ID
	v.ID = ""
	if actual != id || !reviewClock(v.RecordedAt, v.Origin) {
		return nil, errors.New("invalid review task integrity")
	}
	if err := reviewIntegrity(id, v); err != nil {
		return nil, err
	}
	if err := (CaptureReviewTaskInput{v.Version, v.Name, v.Title, v.Labels}).Validate(); err != nil {
		return nil, err
	}
	v.ID = id
	return &v, nil
}
func evidenceFrom(r *exploration.CapturedRecord) (CaptureReviewEvidence, error) {
	if r == nil || reviewTarget(r.Reference) != nil || r.Receipt.ID != r.Reference.CaptureID || r.Receipt.Question.Kind != "explore" || r.Receipt.Question.Explore == nil || !ID.MatchString(r.RecordDigest) || !ID.MatchString(r.Basis) || !validTime(r.SourceReadAt) || !validTime(r.Receipt.CapturedAt) || r.RetainedRecords < 1 || r.RetainedRecords > 1000 || r.Reference.Index >= r.RetainedRecords {
		return CaptureReviewEvidence{}, errors.New("invalid resolved review evidence")
	}
	n, err := strconv.ParseUint(r.MatchingRecords, 10, 64)
	if err != nil || strconv.FormatUint(n, 10) != r.MatchingRecords {
		return CaptureReviewEvidence{}, errors.New("invalid resolved count")
	}
	e := CaptureReviewEvidence{r.Reference, r.Receipt, r.RecordDigest, r.SourceReadAt, r.Basis, r.ObservationState, r.MatchingRecords, r.RetainedRecords, r.RecordsLimited}
	raw, err := json.Marshal(e)
	if err != nil {
		return e, err
	}
	err = json.Unmarshal(raw, &e)
	return e, err
}
func resolveReview(ref exploration.CaptureRecordRef, resolve CaptureReviewResolver) (*exploration.CapturedRecord, CaptureReviewEvidence, error) {
	if resolve == nil {
		return nil, CaptureReviewEvidence{}, errors.New("capture resolver required")
	}
	r, err := resolve(ref)
	if err != nil {
		return nil, CaptureReviewEvidence{}, err
	}
	e, err := evidenceFrom(r)
	if err != nil {
		return nil, e, err
	}
	sum := sha256.Sum256(r.Record)
	if r.Reference != ref || !json.Valid(r.Record) || hex.EncodeToString(sum[:]) != r.RecordDigest {
		return nil, e, errors.New("resolved record identity mismatch")
	}
	return r, e, nil
}
func decisionInput(v CaptureReviewDecision) CaptureReviewDecisionInput {
	return CaptureReviewDecisionInput{v.TaskID, v.Target, v.Parents, v.Judgment, v.Review, v.State, v.Note}
}
func labelIn(task *CaptureReviewTask, j CaptureReviewJudgment) bool {
	return j.Kind != "label" || slices.ContainsFunc(task.Labels, func(l CaptureReviewLabel) bool { return l.Key == j.Value })
}
func (s Store) CreateCaptureReviewDecision(in CaptureReviewDecisionInput, resolve CaptureReviewResolver) (*CaptureReviewDecision, error) {
	return s.createCaptureReviewDecisionAt(in, resolve, time.Now().UTC().Format(time.RFC3339Nano))
}
func (s Store) createCaptureReviewDecisionAt(in CaptureReviewDecisionInput, resolve CaptureReviewResolver, at string) (*CaptureReviewDecision, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	task, err := s.LoadCaptureReviewTask(in.TaskID)
	if err != nil {
		return nil, err
	}
	if !labelIn(task, in.Judgment) {
		return nil, errors.New("label outside pinned task vocabulary")
	}
	_, e, err := resolveReview(in.Target, resolve)
	if err != nil {
		return nil, err
	}
	parents, err := s.collectReview(in.Parents)
	if err != nil {
		return nil, err
	}
	if len(parents) >= 64 {
		return nil, errors.New("review ancestry exceeds 64 references including new decision")
	}
	for _, p := range parents {
		if p.TaskID != in.TaskID || p.Target != in.Target || p.Evidence.RecordDigest != e.RecordDigest {
			return nil, errors.New("review parent belongs to another task or occurrence")
		}
	}
	v := CaptureReviewDecision{Version: 1, TaskID: in.TaskID, Target: in.Target, Parents: slices.Clone(in.Parents), Judgment: in.Judgment, Review: in.Review, State: in.State, Note: in.Note, Evidence: e, RecordedAt: at, Origin: captureReviewOrigin}
	id, err := digest(v)
	if err != nil {
		return nil, err
	}
	v.ID = id
	if err = s.publish("capture-review-decisions", id, v); err != nil {
		return nil, err
	}
	return s.LoadCaptureReviewDecision(id)
}
func (s Store) LoadCaptureReviewDecision(id string) (*CaptureReviewDecision, error) {
	return s.loadReviewDecision(id, map[string]bool{}, new(int), map[string]*CaptureReviewDecision{})
}
func (s Store) loadReviewDecision(id string, path map[string]bool, count *int, loaded map[string]*CaptureReviewDecision) (*CaptureReviewDecision, error) {
	if v := loaded[id]; v != nil {
		return v, nil
	}
	*count++
	if *count > 64 || path[id] {
		return nil, errors.New("review ancestry limit or cycle")
	}
	path[id] = true
	defer delete(path, id)
	var v CaptureReviewDecision
	if err := s.readReview("capture-review-decisions", id, &v); err != nil {
		return nil, err
	}
	actual := v.ID
	v.ID = ""
	if actual != id || v.Version != 1 || !reviewClock(v.RecordedAt, v.Origin) {
		return nil, errors.New("invalid decision integrity")
	}
	if err := reviewIntegrity(id, v); err != nil {
		return nil, err
	}
	v.ID = id
	if err := decisionInput(v).Validate(); err != nil {
		return nil, err
	}
	task, err := s.LoadCaptureReviewTask(v.TaskID)
	if err != nil {
		return nil, err
	}
	if !labelIn(task, v.Judgment) {
		return nil, errors.New("decision outside vocabulary")
	}
	e, err := evidenceFrom(&exploration.CapturedRecord{Reference: v.Evidence.Reference, Receipt: v.Evidence.Receipt, RecordDigest: v.Evidence.RecordDigest, SourceReadAt: v.Evidence.SourceReadAt, Basis: v.Evidence.Basis, ObservationState: v.Evidence.ObservationState, MatchingRecords: v.Evidence.MatchingRecords, RetainedRecords: v.Evidence.RetainedRecords, RecordsLimited: v.Evidence.RecordsLimited})
	if err != nil || e.Reference != v.Target {
		return nil, errors.New("decision evidence differs")
	}
	for _, parent := range v.Parents {
		p, err := s.loadReviewDecision(parent, path, count, loaded)
		if err != nil {
			return nil, err
		}
		if p.TaskID != v.TaskID || p.Target != v.Target || p.Evidence.RecordDigest != v.Evidence.RecordDigest {
			return nil, errors.New("foreign review ancestor")
		}
	}
	loaded[id] = &v
	return &v, nil
}

func (s Store) collectReview(ids []string) (map[string]*CaptureReviewDecision, error) {
	loaded := map[string]*CaptureReviewDecision{}
	count := 0
	for _, id := range ids {
		if _, err := s.loadReviewDecision(id, map[string]bool{}, &count, loaded); err != nil {
			return nil, err
		}
	}
	return loaded, nil
}
func ancestorReview(id, want string, all map[string]*CaptureReviewDecision) bool {
	stack := []string{id}
	seen := map[string]bool{}
	for len(stack) > 0 {
		last := len(stack) - 1
		next := stack[last]
		stack = stack[:last]
		if seen[next] {
			continue
		}
		seen[next] = true
		if next == want {
			return true
		}
		v := all[next]
		if v == nil || len(seen) > 64 {
			return false
		}
		stack = append(stack, v.Parents...)
	}
	return false
}
func eligibleReview(v *CaptureReviewDecision) bool {
	return v.Judgment.Kind == "label" && v.Review == "reviewed" && v.State == "active"
}
func buildReviewManifest(in CaptureReviewManifestInput, all map[string]*CaptureReviewDecision) ([]CaptureReviewManifestItem, error) {
	items := make([]CaptureReviewManifestItem, 0, len(in.Selected))
	seen := map[exploration.CaptureRecordRef]bool{}
	captureID := ""
	for _, id := range in.Selected {
		v := all[id]
		if v == nil || v.TaskID != in.TaskID || !eligibleReview(v) || seen[v.Target] {
			return nil, errors.New("choose one eligible reviewed label per occurrence")
		}
		if captureID != "" && captureID != v.Target.CaptureID {
			return nil, errors.New("one capture per manifest")
		}
		captureID = v.Target.CaptureID
		seen[v.Target] = true
		for _, headID := range in.KnownHeads {
			h := all[headID]
			if h == nil || h.TaskID != in.TaskID {
				return nil, errors.New("known head belongs to another task")
			}
			if h.Target == v.Target {
				if h.Evidence.RecordDigest != v.Evidence.RecordDigest || !ancestorReview(id, headID, all) {
					return nil, errors.New("unresolved pinned known head; select explicit resolution")
				}
			}
		}
		items = append(items, CaptureReviewManifestItem{id, v.Judgment.Value, v.Evidence})
	}
	for _, head := range in.KnownHeads {
		if !seen[all[head].Target] {
			return nil, errors.New("known head outside selected targets")
		}
	}
	return items, nil
}
func (s Store) CreateCaptureReviewManifest(in CaptureReviewManifestInput, resolve CaptureReviewResolver) (*CaptureReviewManifest, error) {
	return s.createCaptureReviewManifestAt(in, resolve, time.Now().UTC().Format(time.RFC3339Nano))
}
func (s Store) createCaptureReviewManifestAt(in CaptureReviewManifestInput, resolve CaptureReviewResolver, at string) (*CaptureReviewManifest, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.LoadCaptureReviewTask(in.TaskID); err != nil {
		return nil, err
	}
	all, err := s.collectReview(append(slices.Clone(in.Selected), in.KnownHeads...))
	if err != nil {
		return nil, err
	}
	items, err := buildReviewManifest(in, all)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		_, e, err := resolveReview(item.Evidence.Reference, resolve)
		if err != nil {
			return nil, err
		}
		if !sameReviewJSON(e, item.Evidence) {
			return nil, errors.New("captured evidence changed before manifest")
		}
	}
	revisions := make([]string, 0, len(all))
	for id := range all {
		revisions = append(revisions, id)
	}
	slices.Sort(revisions)
	v := CaptureReviewManifest{Version: 1, TaskID: in.TaskID, Selected: slices.Clone(in.Selected), KnownHeads: slices.Clone(in.KnownHeads), Historical: true, Revisions: revisions, Items: items, RecordedAt: at, Origin: captureReviewOrigin, Mode: "pinned-review-branches"}
	id, err := digest(v)
	if err != nil {
		return nil, err
	}
	v.ID = id
	if err = s.publish("capture-review-manifests", id, v); err != nil {
		return nil, err
	}
	return s.LoadCaptureReviewManifest(id)
}
func sameReviewJSON(a, b any) bool {
	x, e := json.Marshal(a)
	y, f := json.Marshal(b)
	return e == nil && f == nil && bytes.Equal(x, y)
}
func (s Store) LoadCaptureReviewManifest(id string) (*CaptureReviewManifest, error) {
	var v CaptureReviewManifest
	if err := s.readReview("capture-review-manifests", id, &v); err != nil {
		return nil, err
	}
	actual := v.ID
	v.ID = ""
	if actual != id || v.Version != 1 || v.Mode != "pinned-review-branches" || !reviewClock(v.RecordedAt, v.Origin) {
		return nil, errors.New("invalid manifest integrity")
	}
	if err := reviewIntegrity(id, v); err != nil {
		return nil, err
	}
	v.ID = id
	in := CaptureReviewManifestInput{v.TaskID, v.Selected, v.KnownHeads, v.Historical}
	if err := in.Validate(); err != nil {
		return nil, err
	}
	all, err := s.collectReview(append(slices.Clone(v.Selected), v.KnownHeads...))
	if err != nil {
		return nil, err
	}
	items, err := buildReviewManifest(in, all)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(all))
	for key := range all {
		ids = append(ids, key)
	}
	slices.Sort(ids)
	if !slices.Equal(ids, v.Revisions) || !sameReviewJSON(items, v.Items) {
		return nil, errors.New("manifest pinned ancestry or evidence differs")
	}
	return &v, nil
}
func (s Store) ExportCaptureReviewManifest(id string, resolve CaptureReviewResolver) ([]byte, error) {
	v, err := s.LoadCaptureReviewManifest(id)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	for _, item := range v.Items {
		record, e, err := resolveReview(item.Evidence.Reference, resolve)
		if err != nil {
			return nil, err
		}
		if !sameReviewJSON(e, item.Evidence) {
			return nil, errors.New("export evidence differs from pinned manifest")
		}
		line, err := json.Marshal(struct {
			TaskID       string                       `json:"taskID"`
			DecisionID   string                       `json:"decisionID"`
			Target       exploration.CaptureRecordRef `json:"target"`
			RecordDigest string                       `json:"recordDigest"`
			Label        string                       `json:"label"`
			Record       json.RawMessage              `json:"record"`
		}{v.TaskID, item.DecisionID, item.Evidence.Reference, item.Evidence.RecordDigest, item.Label, record.Record})
		if err != nil {
			return nil, err
		}
		if out.Len()+len(line)+1 > CaptureReviewExportMaxBytes {
			return nil, fmt.Errorf("review export exceeds %d bytes", CaptureReviewExportMaxBytes)
		}
		out.Write(line)
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}
