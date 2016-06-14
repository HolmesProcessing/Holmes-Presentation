package results

import (
	"encoding/json"
	"strings"
	"time"

	"git.hcr.io/cynexit/holmes-api/context"

	"github.com/gocql/gocql"
)

type Result struct {
	Id                string    `json:"id"`
	SHA256            string    `json:"sha256"`
	SchemaVersion     string    `json:"schema_version"`
	UserId            string    `json:"user_id"`
	SourceId          []string  `json:"source_id"`
	SourceTag         []string  `json:"source_tag"`
	ServiceName       string    `json:"service_name"`
	ServiceVersion    string    `json:"service_version"`
	ServiceConfig     string    `json:"service_config"`
	ObjectCategory    []string  `json:"object_category"`
	ObjectType        string    `json:"object_type"`
	Results           string    `json:"results"`
	Tags              []string  `json:"tags"`
	StartedDateTime   time.Time `json:"started_date_time"`
	FinishedDateTime  time.Time `json:"finished_date_time"`
	WatchguardStatus  string    `json:"watchguard_status"`
	WatchguardLog     []string  `json:"watchguard_log"`
	WatchguardVersion string    `json:"watchguard_version"`
}

func GetRoutes() map[string]func(*context.Ctx, *json.RawMessage) *context.Response {
	r := make(map[string]func(*context.Ctx, *json.RawMessage) *context.Response)

	r["get"] = Get
	r["search"] = Search

	return r
}

type GetParameters struct {
	Id string `json:"id"`
}

func Get(c *context.Ctx, parametersRaw *json.RawMessage) *context.Response {
	p := &GetParameters{}
	err := json.Unmarshal(*parametersRaw, p)
	if err != nil {
		return &context.Response{Error: err.Error()}
	}

	result := &Result{}

	uuid, err := gocql.ParseUUID(p.Id)
	if err != nil {
		return &context.Response{Error: err.Error()}
	}

	// TODO: fix results, make everything lower case and revisit this statement
	err = c.C.Query(`SELECT id, sha256, schema_version, user_id, source_id, source_tag, service_name, service_version, service_config, object_category, object_type, results, tags, started_date_time, finished_date_time, watchguard_status, watchguard_log, watchguard_version FROM results WHERE id = ?`, uuid).Scan(
		&result.Id,
		&result.SHA256,
		&result.SchemaVersion,
		&result.UserId,
		&result.SourceId,
		&result.SourceTag,
		&result.ServiceName,
		&result.ServiceVersion,
		&result.ServiceConfig,
		&result.ObjectCategory,
		&result.ObjectType,
		&result.Results,
		&result.Tags,
		&result.StartedDateTime,
		&result.FinishedDateTime,
		&result.WatchguardStatus,
		&result.WatchguardLog,
		&result.WatchguardVersion,
	)

	if err != nil {
		return &context.Response{Error: err.Error()}
	}

	return &context.Response{Result: struct {
		Result *Result
	}{
		result,
	},
	}
}

type SearchParameters struct {
	SHA256      string `json:"sha256"`
	ServiceName string `json:"service_name"`
}

func Search(c *context.Ctx, parametersRaw *json.RawMessage) *context.Response {
	p := &SearchParameters{}
	err := json.Unmarshal(*parametersRaw, p)
	if err != nil {
		return &context.Response{Error: err.Error()}
	}

	if p.SHA256 == "" {
		return &context.Response{Error: "SHA256 is needed!"}
	}

	// since there is only an index on the SHA256
	// we need to cycle trough everything to filter
	// for other fields...
	results := []*Result{}
	result := &Result{}

	// TODO: fix results, make everything lower case and revisit this statement
	q := c.C.Query("SELECT id, sha256, service_name, tags, started_date_time, finished_date_time FROM results WHERE sha256 = ?", strings.ToUpper(p.SHA256))

	iter := q.Iter()
	for iter.Scan(
		&result.Id,
		&result.SHA256,
		&result.ServiceName,
		&result.Tags,
		&result.StartedDateTime,
		&result.FinishedDateTime,
	) {
		if p.ServiceName != "" && result.ServiceName != p.ServiceName {
			result = &Result{}
			continue
		}

		results = append(results, result)
		result = &Result{}
	}

	if err = iter.Close(); err != nil {
		return &context.Response{Error: err.Error()}
	}

	return &context.Response{Result: struct {
		Results []*Result
	}{
		results,
	},
	}
}
