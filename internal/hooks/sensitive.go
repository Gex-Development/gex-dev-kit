package hooks

import (
	"strings"
	"unicode"
)

const (
	sensitiveCategoryDestructive = "destrutiva"
	sensitiveCategoryPublish     = "publicação externa"
	sensitiveCategorySecret      = "segredo"
	sensitiveCategoryDynamic     = "comando dinâmico"
	sensitiveCategoryTooBig      = "comando grande demais para análise"
	sensitiveRuleDynamic         = "substituição ou eval com palavra sensível"
	sensitiveRuleTooBig          = "entrada acima de 8 MiB ou ilegível"
)

var sensitiveWrappers = map[string]bool{
	"env": true, "sudo": true, "command": true, "exec": true, "nohup": true,
	"time": true, "nice": true, "timeout": true, "xargs": true,
}

var sensitiveOptionArgs = map[string]bool{
	"-u": true, "-g": true, "-C": true, "-U": true, "-n": true, "-I": true,
	"-d": true, "-P": true, "-L": true, "-s": true, "-k": true,
}

var sensitivePublishRules = []string{
	"gh pr merge", "gh release create", "gh release upload", "gh release delete",
	"gh repo create", "gh repo delete", "gh issue create", "npm publish",
	"pnpm publish", "yarn publish", "docker push", "twine upload", "cargo publish",
}

var sensitiveSecretReaders = map[string]bool{
	"cat": true, "less": true, "more": true, "head": true, "tail": true,
	"bat": true, "type": true, "Get-Content": true,
}

var sensitiveKeywords = map[string]bool{
	"rm": true, "git": true, "terraform": true, "kubectl": true, "aws": true,
	"psql": true, "mysql": true, "sqlite3": true, "gh": true, "npm": true,
	"pnpm": true, "yarn": true, "docker": true, "twine": true, "cargo": true,
	"printenv": true, "env": true, "op": true, "security": true, "eval": true,
	"cat": true, "less": true, "more": true, "head": true, "tail": true,
	"bat": true, "type": true, "get-content": true, "remove-item": true,
}

func ClassifySensitive(cmds []Command, tool, filePath string, prePRInstalled bool) Decision {
	if tool == "Read" {
		if sensitiveDotEnvName(filePath) {
			return sensitiveAsk(sensitiveCategorySecret, "leitura de arquivo sensível")
		}
		return Decision{}
	}
	if tool != "Bash" && tool != "PowerShell" {
		return Decision{}
	}
	for _, cmd := range cmds {
		if category, rule, ok := sensitiveMatch(cmd.Words, prePRInstalled); ok {
			return sensitiveAsk(category, rule)
		}
	}
	return Decision{}
}

func CheckSensitive(in AgentInput, scope func(dirs []string) bool, prePRInstalled func() bool) Decision {
	return sensitiveCheck(in, scope, prePRInstalled, ParseShell)
}

func sensitiveCheck(in AgentInput, scope func(dirs []string) bool, prePRInstalled func() bool, parse func(string) (Script, error)) Decision {
	if in.Truncated || !in.Parsed {
		return sensitiveInScope(scope, []string{in.Cwd}, sensitiveAsk(sensitiveCategoryTooBig, sensitiveRuleTooBig))
	}
	if in.Tool == "Read" {
		return sensitiveInScope(scope, []string{in.Cwd}, ClassifySensitive(nil, in.Tool, in.FilePath, false))
	}
	if in.Tool != "Bash" && in.Tool != "PowerShell" {
		return Decision{}
	}
	script, err := parse(in.Command)
	if err != nil {
		return sensitiveInScope(scope, []string{in.Cwd}, sensitiveAsk(sensitiveCategoryTooBig, sensitiveRuleTooBig))
	}
	cmds := sensitiveExpand(script)
	installed := prePRInstalled != nil && sensitiveHasPRCreate(cmds) && prePRInstalled()
	decision := ClassifySensitive(cmds, in.Tool, "", installed)
	return sensitiveInScope(scope, sensitiveDirs(in.Cwd, cmds), decision)
}

func sensitiveInScope(scope func(dirs []string) bool, dirs []string, decision Decision) Decision {
	if !decision.Ask && !decision.Block {
		return Decision{}
	}
	if scope == nil || !scope(dirs) {
		return Decision{}
	}
	return decision
}

func sensitiveAsk(category, rule string) Decision {
	return Decision{Ask: true, Category: category, Rule: rule}
}

func sensitiveExpand(script Script) []Command {
	cmds := make([]Command, 0, len(script.Commands))
	for _, cmd := range script.Commands {
		cmds = append(cmds, cmd)
		cmds = append(cmds, Effective(cmd)...)
	}
	if script.Dynamic {
		for i := range cmds {
			cmds[i].Words = sensitiveMarkDynamic(cmds[i].Words)
		}
	}
	return cmds
}

func sensitiveMarkDynamic(words []Word) []Word {
	marked := make([]Word, len(words))
	for i, w := range words {
		w.Dynamic = true
		marked[i] = w
	}
	return marked
}

func sensitiveDirs(cwd string, cmds []Command) []string {
	dirs := []string{cwd}
	for _, cmd := range cmds {
		if cmd.Dir != "" {
			dirs = append(dirs, cmd.Dir)
		}
	}
	return dirs
}

func sensitiveHasPRCreate(cmds []Command) bool {
	for _, args := range sensitiveCandidates(cmds) {
		if sensitiveHasPrefix(args, "gh", "pr", "create") {
			return true
		}
	}
	return false
}

func sensitiveCandidates(cmds []Command) [][]string {
	var out [][]string
	for _, cmd := range cmds {
		out = append(out, sensitiveValues(cmd.Words), sensitiveValues(sensitiveUnwrap(cmd.Words)))
	}
	return out
}

func sensitiveValues(words []Word) []string {
	values := make([]string, 0, len(words))
	for _, w := range words {
		values = append(values, w.Value)
	}
	return values
}

func sensitiveUnwrap(words []Word) []Word {
	for len(words) > 0 && sensitiveWrappers[words[0].Value] {
		name := words[0].Value
		words = words[1:]
		for len(words) > 0 {
			value := words[0].Value
			if sensitiveOptionArgs[value] && len(words) > 1 {
				words = words[2:]
			} else if strings.HasPrefix(value, "-") || (name == "env" && strings.Contains(value, "=")) {
				words = words[1:]
			} else {
				break
			}
		}
		if name == "timeout" && len(words) > 0 {
			words = words[1:]
		}
	}
	return words
}

func sensitiveMatch(words []Word, prePRInstalled bool) (string, string, bool) {
	for _, args := range [][]string{sensitiveValues(words), sensitiveValues(sensitiveUnwrap(words))} {
		if category, rule, ok := sensitiveStatic(args, prePRInstalled); ok {
			return category, rule, true
		}
	}
	if sensitiveDynamic(words) {
		return sensitiveCategoryDynamic, sensitiveRuleDynamic, true
	}
	return "", "", false
}

func sensitiveStatic(args []string, prePRInstalled bool) (string, string, bool) {
	if len(args) == 0 {
		return "", "", false
	}
	if category, rule, ok := sensitiveDestructiveRule(args); ok {
		return category, rule, true
	}
	if rule, ok := sensitivePublishRule(args); ok {
		return sensitiveCategoryPublish, rule, true
	}
	if !prePRInstalled && sensitiveHasPrefix(args, "gh", "pr", "create") {
		return sensitiveCategoryPublish, "gh pr create", true
	}
	if rule, ok := sensitiveSecretRule(args); ok {
		return sensitiveCategorySecret, rule, true
	}
	return "", "", false
}

func sensitiveDestructiveRule(args []string) (string, string, bool) {
	category := sensitiveCategoryDestructive
	switch {
	case sensitiveRecursiveRemove(args):
		return category, "rm -r em alvo amplo", true
	case sensitiveHasPrefix(args, "git", "reset") && sensitiveHasArg(args, "--hard"):
		return category, "git reset --hard", true
	case sensitiveHasPrefix(args, "git", "clean") && sensitiveHasShortFlag(args, 'f'):
		return category, "git clean -f", true
	case sensitiveHasPrefix(args, "git", "push") && sensitiveForceArg(args[2:]):
		return category, "git push --force", true
	case sensitiveHasPrefix(args, "git", "branch") && sensitiveHasArg(args, "-D"):
		return category, "git branch -D", true
	case sensitiveHasPrefix(args, "git", "checkout") && sensitiveHasArg(args, "--") && sensitiveHasArg(args, "."):
		return category, "git checkout -- .", true
	case sensitiveHasPrefix(args, "git", "restore") && sensitiveHasArg(args, ".") && !sensitiveHasArg(args, "--staged"):
		return category, "git restore .", true
	case sensitiveHasPrefix(args, "terraform", "destroy"):
		return category, "terraform destroy", true
	case sensitiveHasPrefix(args, "terraform", "apply"):
		return category, "terraform apply", true
	case sensitiveHasPrefix(args, "kubectl", "delete"):
		return category, "kubectl delete", true
	case sensitiveHasPrefix(args, "aws", "s3", "rm") && sensitiveHasArg(args, "--recursive"):
		return category, "aws s3 rm --recursive", true
	case sensitiveHasPrefix(args, "Remove-Item") && sensitiveHasFoldFlag(args, "-recurse"):
		return category, "Remove-Item -Recurse", true
	case sensitiveDBDrop(args):
		return category, "DROP TABLE, DROP DATABASE ou TRUNCATE", true
	}
	return "", "", false
}

func sensitivePublishRule(args []string) (string, bool) {
	if sensitiveHasPrefix(args, "git", "push") {
		return "git push", true
	}
	for _, rule := range sensitivePublishRules {
		if sensitiveHasPrefix(args, strings.Fields(rule)...) {
			return rule, true
		}
	}
	return "", false
}

func sensitiveSecretRule(args []string) (string, bool) {
	switch {
	case sensitiveHasPrefix(args, "printenv"):
		return "printenv", true
	case sensitiveHasPrefix(args, "env") && len(args) == 1:
		return "env sem comando", true
	case sensitiveHasPrefix(args, "gh", "auth", "token"):
		return "gh auth token", true
	case sensitiveHasPrefix(args, "aws", "configure", "get"):
		return "aws configure get", true
	case sensitiveHasPrefix(args, "security") && sensitiveHasArg(args, "find-generic-password") && sensitiveHasArg(args, "-w"):
		return "security find-generic-password -w", true
	case sensitiveHasPrefix(args, "op", "read"):
		return "op read", true
	case sensitiveHasPrefix(args, "echo") && sensitiveSecretVarArg(args[1:]):
		return "echo de variável sensível", true
	case sensitiveSecretReaders[args[0]] && sensitiveDotEnvArg(args[1:]):
		return "leitura de arquivo sensível", true
	}
	return "", false
}

func sensitiveDynamic(words []Word) bool {
	if !sensitiveHasDynamic(words) {
		return false
	}
	for _, w := range words {
		for _, token := range sensitiveTokens(w.Value) {
			if sensitiveKeywords[strings.ToLower(token)] || sensitiveSecretName(token) || sensitiveDotEnvName(token) {
				return true
			}
		}
	}
	return false
}

func sensitiveHasDynamic(words []Word) bool {
	for _, w := range words {
		if w.Dynamic {
			return true
		}
	}
	return false
}

func sensitiveTokens(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' && r != '.'
	})
}

func sensitiveHasPrefix(args []string, prefix ...string) bool {
	if len(args) < len(prefix) {
		return false
	}
	for i, p := range prefix {
		if !strings.EqualFold(args[i], p) {
			return false
		}
	}
	return true
}

func sensitiveHasArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func sensitiveHasFoldFlag(args []string, flag string) bool {
	for _, arg := range args {
		if strings.HasPrefix(strings.ToLower(arg), flag) {
			return true
		}
	}
	return false
}

func sensitiveHasShortFlag(args []string, letter rune) bool {
	for _, arg := range args {
		if arg == "--force" {
			return true
		}
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && strings.ContainsRune(arg, letter) {
			return true
		}
	}
	return false
}

func sensitiveForceArg(args []string) bool {
	for _, arg := range args {
		if arg == "-f" || strings.HasPrefix(arg, "--force") || strings.HasPrefix(arg, "+") {
			return true
		}
	}
	return false
}

func sensitiveRecursiveRemove(args []string) bool {
	if !sensitiveHasPrefix(args, "rm") {
		return false
	}
	recursive, dangerous := false, false
	for _, arg := range args[1:] {
		switch {
		case arg == "--recursive":
			recursive = true
		case strings.HasPrefix(arg, "--"):
		case strings.HasPrefix(arg, "-") && strings.ContainsAny(arg, "rR"):
			recursive = true
		case arg == "/" || arg == "~" || arg == "*" || arg == "." || arg == "..":
			dangerous = true
		}
	}
	return recursive && dangerous
}

func sensitiveDBDrop(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch strings.ToLower(args[0]) {
	case "psql", "mysql", "sqlite3":
	default:
		return false
	}
	text := strings.ToUpper(strings.Join(args[1:], " "))
	return strings.Contains(text, "DROP TABLE") || strings.Contains(text, "DROP DATABASE") || strings.Contains(text, "TRUNCATE")
}

func sensitiveSecretVarArg(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "$") && sensitiveSecretName(strings.Trim(arg, "${}")) {
			return true
		}
	}
	return false
}

func sensitiveSecretName(name string) bool {
	upper := strings.ToUpper(name)
	for _, keyword := range []string{"TOKEN", "SECRET", "KEY", "PASSWORD"} {
		if strings.Contains(upper, keyword) {
			return true
		}
	}
	return false
}

func sensitiveDotEnvArg(args []string) bool {
	for _, arg := range args {
		if sensitiveDotEnvName(arg) {
			return true
		}
	}
	return false
}

func sensitiveDotEnvName(path string) bool {
	base := strings.ToLower(path[strings.LastIndexAny(path, "/"+string(rune(92)))+1:])
	if strings.HasPrefix(base, ".env") || strings.HasPrefix(base, "id_rsa") || base == "credentials" {
		return true
	}
	return strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key")
}
