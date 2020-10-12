package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/mattn/go-isatty"
	"github.com/peterh/liner"
	"github.com/y-bash/go-tran"
	"github.com/y-bash/go-tran/config"
)

const version = "1.0.1"

var cfg *config.Config

func isTerminal(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func helpToNonTerm() {
	msg := `GO-TRAN (The language translator), version %s

usage:  tran [option...] [file...]

options:
    -h          show summary of options.
    -l          list the language codes(ISO639-1).
    -s CODE     specify the source language with CODE(ISO639-1).
    -t CODE     specify the target language with CODE(ISO639-1).
    -v          output version information.
`
	fmt.Fprintf(os.Stderr, msg, version)
}

func helpToTerm() {
	text := `┌──┬──────────┬────────┐
│Cmd │    Description     │    Examples    │
├──┼──────────┼──┬─────┤
│ h  │Show help           │h   │          │
│ l  │Show language codes │l en│l nor     │
│ s  │Source language code│s en│s french  │
│ t  │Target language code│t ja│t italian │
│ q  │Quit                │q   │          │
└──┴──────────┴──┴─────┘ `

	fmt.Fprintln(os.Stderr, cfg.InfoColor.Apply(text))
}

func langCodesToNonTerm(w io.Writer) {
	text := `Code Language name
---- -------------
{{range .}} {{.Code}}  {{.Name}}
{{end -}}
`
	a := tran.AllLangList()
	tmpl := template.Must(template.New("lang").Parse(text))
	tmpl.Execute(w, a)
}

func langCodesToTerm(w io.Writer, substr string) (ok bool) {
	text := `┌──┬──────────┐
│Code│Language name       │
├──┼──────────┤
{{range .}}│ {{.Code}} │{{printf "%-20s" .Name}}│
{{end -}}
└──┴──────────┘ 
`
	a := cfg.APIEndpoint.LangListContains(substr)
	if len(a) == 0 {
		return false
	}
	tmpl := template.Must(template.New("lang").Parse(text))
	var buf bytes.Buffer
	tmpl.Execute(&buf, a)
	fmt.Fprint(w, cfg.InfoColor.Apply(string(buf.Bytes())))
	return true
}

func brackets(s string) string {
	if s == "" {
		return ""
	}
	return "(" + s + ")"
}

func commandLangCodes(in string) {
	if in != "l" {
		in = in[2:]
	}
	if !langCodesToTerm(os.Stderr, in) {
		msg := cfg.ErrorColor.Apply("%q is not found\n")
		fmt.Fprintf(os.Stderr, msg, in)
	}
}

func commandSource(in, curr string) (source string, ok bool) {
	var code, name string
	if in == "s" {
		code = cfg.DefaultSourceCode
		name = cfg.DefaultSourceName
		ok = true
	} else {
		if strings.HasPrefix(in, "s ") {
			in = strings.TrimSpace(string([]rune(in)[2:]))
		}
		if code, name, ok = cfg.APIEndpoint.LookupLang(in); !ok {
			code, name, ok = tran.LookupPlang(in)
		}
		if !ok {
			msg := cfg.ErrorColor.Apply("%q is not found\n")
			fmt.Fprintf(os.Stderr, msg, in)
			return "", ok
		}
	}
	if curr != code {
		msg := cfg.StateColor.Apply("Srouce changed: %s %s\n")
		fmt.Fprintf(os.Stderr, msg, name, brackets(code))
	}
	return code, ok
}

func commandTarget(in, curr string) (target string, ok bool) {
	var code, name string
	if in == "t" {
		code = cfg.DefaultTargetCode
		name = cfg.DefaultTargetName
		ok = true
	} else {
		if strings.HasPrefix(in, "t ") {
			in = strings.TrimSpace(string([]rune(in)[2:]))
		}
		if code, name, ok = cfg.APIEndpoint.LookupLang(in); !ok {
			code, name, ok = tran.LookupPlang(in)
		}
		if !ok {
			msg := cfg.ErrorColor.Apply("%q is not found\n")
			fmt.Fprintf(os.Stderr, msg, in)
			return "", ok
		}
	}
	if curr != code {
		msg := cfg.StateColor.Apply("Target changed: %s %s\n")
		fmt.Fprintf(os.Stderr, msg, name, brackets(code))
	}
	return code, ok
}

func interact() {
	fmt.Fprintf(os.Stderr, "Welcome to the GO-TRAN! (Ver %s)\n", version)
	helpToTerm()
	source := cfg.DefaultSourceCode
	target := cfg.DefaultTargetCode
	line := liner.NewLiner()
	defer line.Close()
	for {
		pr := fmt.Sprintf("%s:%s> ", source, target)
		in, err := line.Prompt(pr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if len(in) <= 0 {
			continue
		}
		in = strings.TrimSpace(in)
		switch {
		case in == "q":
			fmt.Fprintln(os.Stderr, "Leaving GO-TRAN.")
			return

		case in == "h":
			helpToTerm()

		case in == "l" || strings.HasPrefix(in, "l "):
			commandLangCodes(in)

		case in == "s" || strings.HasPrefix(in, "s "):
			if code, ok := commandSource(in, source); ok {
				source = code
			}
		case len(in) <= 2 || strings.HasPrefix(in, "t "):
			if code, ok := commandTarget(in, target); ok {
				target = code
			}
		default:
			if out, ok := tran.Ptranslate(in, target); ok {
				fmt.Fprintln(os.Stderr, cfg.ResultColor.Apply(out))
			} else {
				out, err := cfg.APIEndpoint.Translate(in, source, target)
				if err != nil {
					fmt.Fprintln(os.Stderr, cfg.ErrorColor.Apply(err.Error()))
				} else {
					fmt.Fprintln(os.Stderr, cfg.ResultColor.Apply(out))
				}
			}
		}
		line.AppendHistory(in)
	}
}

func scanText(sc *bufio.Scanner, n int) (out string, eof bool) {
	var sb strings.Builder
	sb.Grow(4096)
	for i := 0; i < n; {
		if !sc.Scan() {
			break
		}
		s := sc.Text()
		sb.WriteString(s)
		sb.WriteString("\n")
		i += len([]rune(s))
	}
	out = sb.String()
	return out, len(out) == 0
}

func translate(w io.Writer, r io.Reader) error {
	source := cfg.DefaultSourceCode
	target := cfg.DefaultTargetCode
	tran := cfg.APIEndpoint.Translate
	limit := cfg.APILimitNChars
	sc := bufio.NewScanner(r)
	for {
		in, eof := scanText(sc, limit)
		if eof {
			break
		}
		out, err := tran(in, source, target)
		if err != nil {
			return err
		}
		fmt.Fprint(w, out)
	}
	return nil
}

func batch(paths []string) error {
	if len(paths) == 0 {
		translate(os.Stdout, os.Stdin)
		return nil
	}
	for _, path := range paths {
		var f *os.File
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		translate(os.Stdout, f)
	}
	return nil
}

func main() {
	var help, lang, ver bool
	var source, target string

	flag.Usage	= helpToNonTerm
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&lang, "l", false, "Show language codes (ISO-639-1)")
	flag.StringVar(&source, "s", "", "Source language code")
	flag.StringVar(&target, "t", "", "Target language code")
	flag.BoolVar(&ver, "v", false, "Show version")
	flag.Parse()

	if help {
		flag.Usage()
		return
	}
	if lang {
		langCodesToNonTerm(os.Stdout)
		return
	}
	if ver {
		fmt.Fprintf(os.Stderr, "GO-TRAN Version %s\n", version)
		return
	}

	var err error
	if cfg, err = config.Load(source, target); err != nil {
		fmt.Fprintf(os.Stderr, "GO-TRAN: %s\n", err)
		os.Exit(1)
	}
	if flag.NArg() == 0 && isTerminal(os.Stdin.Fd()) {
		interact()
		return
	}
	err = batch(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "GO-TRAN: %s\n", err)
		os.Exit(1)
	}
}
