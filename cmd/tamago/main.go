package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type AW struct {
	Version     string   `yaml:"version" json:"version"`
	Name        string   `yaml:"name" json:"name"`
	Goal        string   `yaml:"goal" json:"goal"`
	Context     string   `yaml:"context,omitempty" json:"context,omitempty"`
	Prompt      string   `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Constraints []string `yaml:"constraints,omitempty" json:"constraints,omitempty"`
	Tools       []string `yaml:"tools,omitempty" json:"tools,omitempty"`
	Success     []string `yaml:"success" json:"success"`
	Examples    []string `yaml:"examples,omitempty" json:"examples,omitempty"`
}

type WF struct {
	Version string `yaml:"version" json:"version"`
	Name    string `yaml:"name" json:"name"`
	Goal    string `yaml:"goal" json:"goal"`
	Steps   []Step `yaml:"steps" json:"steps"`
}

type Step struct {
	ID        string   `yaml:"id" json:"id"`
	Kind      string   `yaml:"kind" json:"kind"`
	Prompt    string   `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Tool      string   `yaml:"tool,omitempty" json:"tool,omitempty"`
	Command   string   `yaml:"command,omitempty" json:"command,omitempty"`
	Condition string   `yaml:"condition,omitempty" json:"condition,omitempty"`
	Artifact  string   `yaml:"artifact,omitempty" json:"artifact,omitempty"`
	Needs     []string `yaml:"needs,omitempty" json:"needs,omitempty"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Temp     float64   `json:"temperature,omitempty"`
}
type Message struct { Role string `json:"role"`; Content string `json:"content"` }
type ChatResponse struct { Choices []struct { Message Message `json:"message"` } `json:"choices"` }

func main() {
	if len(os.Args) < 2 { usage(); os.Exit(2) }
	switch os.Args[1] {
	case "validate": validateCmd(os.Args[2:])
	case "compile": compileCmd(os.Args[2:])
	default: usage(); os.Exit(2)
	}
}

func usage() { fmt.Println("tamago — AW YAML → WF YAML compiler\n\nUsage:\n  tamago validate <file.aw.yaml>\n  tamago compile <file.aw.yaml> [-o file.wf.yaml] [--dry-run]\n\nEnvironment:\n  TAMAGO_BASE_URL  OpenAI-compatible API base (default http://127.0.0.1:1234/v1)\n  TAMAGO_API_KEY   API key (optional for LM Studio)\n  TAMAGO_MODEL     model name (default local-model)") }

func validateCmd(args []string) {
	if len(args) != 1 { fmt.Fprintln(os.Stderr, "validate requires one AW file"); os.Exit(2) }
	aw, err := loadAW(args[0]); if err != nil { fail(err) }
	if err := validateAW(aw); err != nil { fail(err) }
	fmt.Printf("valid AW: %s\n", aw.Name)
}

func compileCmd(args []string) {
	fs := flag.NewFlagSet("compile", flag.ExitOnError)
	out := fs.String("o", "", "output WF YAML")
	dry := fs.Bool("dry-run", false, "compile using the built-in example planner")
	fs.Parse(args)
	if fs.NArg() != 1 { fail(errors.New("compile requires one AW file")) }
	input := fs.Arg(0)
	aw, err := loadAW(input); if err != nil { fail(err) }
	if err := validateAW(aw); err != nil { fail(err) }

	var wf WF
	if *dry { wf = dryRun(aw) } else { wf, err = plan(aw); if err != nil { fail(err) } }
	if err := validateWF(wf); err != nil { fail(fmt.Errorf("generated WF rejected: %w", err)) }
	data, err := yaml.Marshal(wf); if err != nil { fail(err) }
	if *out == "" { fmt.Print(string(data)); return }
	if err := os.WriteFile(*out, data, 0644); err != nil { fail(err) }
	fmt.Printf("compiled: %s -> %s\n", input, *out)
}

func loadAW(path string) (AW, error) { b, err := os.ReadFile(path); if err != nil { return AW{}, err }; var aw AW; if err := yaml.Unmarshal(b, &aw); err != nil { return AW{}, err }; return aw, nil }

func validateAW(a AW) error {
	if a.Version == "" || a.Name == "" || a.Goal == "" { return errors.New("AW requires version, name, goal") }
	if len(a.Success) == 0 { return errors.New("AW requires at least one success criterion") }
	return nil
}
func validateWF(w WF) error {
	if w.Version == "" || w.Name == "" || w.Goal == "" { return errors.New("WF requires version, name, goal") }
	if len(w.Steps) == 0 { return errors.New("WF requires steps") }
	seen := map[string]bool{}
	allowed := map[string]bool{"agent":true,"tool":true,"command":true,"condition":true,"artifact":true}
	for _, s := range w.Steps {
		if s.ID == "" || !allowed[s.Kind] { return fmt.Errorf("invalid step %q", s.ID) }
		if seen[s.ID] { return fmt.Errorf("duplicate step id: %s", s.ID) }; seen[s.ID] = true
	}
	return nil
}

func dryRun(a AW) WF { return WF{Version:a.Version, Name:a.Name, Goal:a.Goal, Steps:[]Step{
	{ID:"observe", Kind:"agent", Prompt:a.Prompt},
	{ID:"research", Kind:"tool", Tool:first(a.Tools, "research") , Needs:[]string{"observe"}},
	{ID:"verify", Kind:"condition", Condition:strings.Join(a.Success, " && "), Needs:[]string{"research"}},
}} }
func first(xs []string, fallback string) string { if len(xs)>0 { return xs[0] }; return fallback }

func plan(a AW) (WF, error) {
	base := getenv("TAMAGO_BASE_URL", "http://127.0.0.1:1234/v1")
	model := getenv("TAMAGO_MODEL", "local-model")
	prompt := buildPrompt(a)
	body := ChatRequest{Model:model, Temp:0, Messages:[]Message{{Role:"system",Content:"You are tamago, a compiler. Return ONLY valid JSON matching the WF object: version,name,goal,steps. Each step has id and kind; kind must be agent, tool, command, condition, or artifact. Do not invent tools outside the AW tools unless a command is explicitly required."},{Role:"user",Content:prompt}}}
	b,_ := json.Marshal(body)
	req, err := http.NewRequest("POST", strings.TrimRight(base,"/")+"/chat/completions", bytes.NewReader(b)); if err != nil { return WF{}, err }
	req.Header.Set("Content-Type","application/json"); if key:=os.Getenv("TAMAGO_API_KEY"); key!="" { req.Header.Set("Authorization","Bearer "+key) }
	resp, err := http.DefaultClient.Do(req); if err != nil { return WF{}, err }; defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body); if resp.StatusCode >= 300 { return WF{}, fmt.Errorf("LLM HTTP %s: %s", resp.Status, raw) }
	var cr ChatResponse; if err:=json.Unmarshal(raw,&cr); err!=nil { return WF{}, err }; if len(cr.Choices)==0 { return WF{}, errors.New("LLM returned no choices") }
	text := stripFence(cr.Choices[0].Message.Content)
	var wf WF; if err:=json.Unmarshal([]byte(text),&wf); err!=nil { return WF{}, fmt.Errorf("LLM output is not WF JSON: %w",err) }
	return wf,nil
}

func buildPrompt(a AW) string { b,_:=json.MarshalIndent(a,"","  "); return "Compile this AW into an executable WF. Preserve the goal and success criteria. Use tools from the declared list. Prefer small deterministic steps and explicit dependencies.\n\nAW:\n"+string(b) }
func stripFence(s string) string { s=strings.TrimSpace(s); s=strings.TrimPrefix(s,"```json"); s=strings.TrimPrefix(s,"```"); s=strings.TrimSuffix(s,"```"); return strings.TrimSpace(s) }
func getenv(k,d string) string { if v:=os.Getenv(k); v!="" { return v }; return d }
func fail(err error) { fmt.Fprintln(os.Stderr,"tamago:",err); os.Exit(1) }
