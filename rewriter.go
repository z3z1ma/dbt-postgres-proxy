package proxy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	uuid "github.com/google/uuid"
)

// RPC protocol
type CompileQueryArgs struct {
	Timeout int    `json:"timeout"`
	SQL     string `json:"sql"`
	Name    string `json:"name"`
}

type CompileQueryRPC struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      string `json:"id"`
	Params  CompileQueryArgs `json:"params"`
}

type PendingRPC struct {
	Result struct {
		RequestToken    string `json:"request_token"`
	} `json:"result"`
	ID      string `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
}

type PollQueryArgs struct {
	RequestToken string `json:"request_token"`
	Logs         bool   `json:"logs"`
	LogsStart    int    `json:"logs_start"`
}

type PollQuery struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      string `json:"id"`
	Params  PollQueryArgs `json:"params"`
}

type CompileQueryResults struct {
	Result struct {
		Results []struct {
			CompiledSQL string        `json:"compiled_sql"`
			GeneratedAt time.Time     `json:"generated_at"`
			Logs        []interface{} `json:"logs"`
			Node        struct {
				Alias     string      `json:"alias"`
				BuildPath interface{} `json:"build_path"`
				Checksum  struct {
					Checksum string `json:"checksum"`
					Name     string `json:"name"`
				} `json:"checksum"`
				Columns struct {
				} `json:"columns"`
				Compiled     bool        `json:"compiled"`
				CompiledPath interface{} `json:"compiled_path"`
				CompiledSQL  string      `json:"compiled_sql"`
				Config       struct {
					Alias       interface{} `json:"alias"`
					ColumnTypes struct {
					} `json:"column_types"`
					Database     interface{} `json:"database"`
					Enabled      bool        `json:"enabled"`
					FullRefresh  interface{} `json:"full_refresh"`
					Materialized string      `json:"materialized"`
					Meta         struct {
					} `json:"meta"`
					OnSchemaChange string `json:"on_schema_change"`
					PersistDocs    struct {
					} `json:"persist_docs"`
					PostHook []interface{} `json:"post-hook"`
					PreHook  []interface{} `json:"pre-hook"`
					Quoting  struct {
					} `json:"quoting"`
					Schema interface{}   `json:"schema"`
					Tags   []interface{} `json:"tags"`
				} `json:"config"`
				CreatedAt int    `json:"created_at"`
				Database  string `json:"database"`
				Deferred  bool   `json:"deferred"`
				DependsOn struct {
					Macros []interface{} `json:"macros"`
					Nodes  []interface{} `json:"nodes"`
				} `json:"depends_on"`
				Description string `json:"description"`
				Docs        struct {
					Show bool `json:"show"`
				} `json:"docs"`
				ExtraCtes         []interface{} `json:"extra_ctes"`
				ExtraCtesInjected bool          `json:"extra_ctes_injected"`
				Fqn               []string      `json:"fqn"`
				Meta              struct {
				} `json:"meta"`
				Name             string        `json:"name"`
				OriginalFilePath string        `json:"original_file_path"`
				PackageName      string        `json:"package_name"`
				PatchPath        interface{}   `json:"patch_path"`
				Path             string        `json:"path"`
				RawSQL           string        `json:"raw_sql"`
				Refs             []interface{} `json:"refs"`
				RelationName     interface{}   `json:"relation_name"`
				ResourceType     string        `json:"resource_type"`
				RootPath         string        `json:"root_path"`
				Schema           string        `json:"schema"`
				Sources          []interface{} `json:"sources"`
				Tags             []interface{} `json:"tags"`
				UniqueID         string        `json:"unique_id"`
			} `json:"node"`
			RawSQL string `json:"raw_sql"`
			Timing []struct {
				CompletedAt time.Time `json:"completed_at"`
				Name        string    `json:"name"`
				StartedAt   time.Time `json:"started_at"`
			} `json:"timing"`
		} `json:"results"`
		GeneratedAt time.Time     `json:"generated_at"`
		ElapsedTime float64       `json:"elapsed_time"`
		Logs        []interface{} `json:"logs"`
		Tags        struct {
			Command string `json:"command"`
			Branch  string `json:"branch"`
		} `json:"tags"`
		Status string `json:"state"`
	} `json:"result"`
	ID      string `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
}

/// Rewriters will be constructed per goroutine, as some of them may have state that isn't safe to share
type QueryRewriterFactory interface {
	Create() (QueryRewriter, error)
}

// Generic query rewriter interface
type QueryRewriter interface {
	RewriteQuery(string) (string, error)
	RewriteParse(string) (string, error)
}

// dbt compiler implementation
type DbtRewriter struct {
	host string
	port int
}

type DbtRewriterFactory struct {
	host string
	port int
}

func NewDbtRewriterFactory(host string, port int) *DbtRewriterFactory {
	return &DbtRewriterFactory{
		host: host,
		port: port,
	}
}

func (r *DbtRewriterFactory) Create() (QueryRewriter, error) {
	return &DbtRewriter{host: r.host, port: r.port}, nil
}

func (r *DbtRewriter) RewriteQuery(query string) (string, error) {
	return r.rewriteInternal(query)
}

func (r *DbtRewriter) RewriteParse(query string) (string, error) {
	return r.rewriteInternal(query)
}

func (r *DbtRewriter) rewriteInternal(query string) (string, error) {

	if (strings.Contains(query, "{{") || strings.Contains(query, "{%")) {
		// Basic yet effective check for Jinja infused query
		serverURL := "http://" + r.host + ":" + fmt.Sprint(r.port) + "/jsonrpc"
		taskId := uuid.New().String()
		args := CompileQueryRPC{
			Jsonrpc: "2.0",
			Method: "compile_sql",
			ID: taskId,
			Params: CompileQueryArgs{
				Timeout: 60,
				SQL: base64.StdEncoding.EncodeToString([]byte(query)),
				Name: "dbt_pg_proxy_query",
			},
		}
		
		body, _ := json.Marshal(args)
		resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Fatal("Error:", err)
		}
		var reply PendingRPC
		json.NewDecoder(resp.Body).Decode(&reply)

		pollingArgs := PollQuery{
			Jsonrpc: "2.0",
			Method: "poll",
			ID: taskId,
			Params: PollQueryArgs{
				RequestToken: reply.Result.RequestToken,
				Logs: true,
				LogsStart: 1,
			},
		}

		body, _ = json.Marshal(pollingArgs)
		var queryResult CompileQueryResults
		var compiledSQL string

		for {
			var result CompileQueryResults
			resp, err = http.Post(serverURL, "application/json", bytes.NewBuffer(body))
			if err != nil { log.Fatal("Error:", err) }
			json.NewDecoder(resp.Body).Decode(&result)
			if result.Result.Status == "running" { 
				time.Sleep(100 * time.Millisecond)
				continue
			}
			queryResult = result
			break
		}
		
		if queryResult.Result.Status == "success" {
			compiledSQL = queryResult.Result.Results[0].CompiledSQL
			query = compiledSQL
			fmt.Println("")
			fmt.Println("====================== dbt output ======================")
			fmt.Println("")
			fmt.Println(query)
		} else {
			return query, fmt.Errorf("failed to compile query with dbt, check for syntax errors")
		}

	}
	
	return query, nil
}
