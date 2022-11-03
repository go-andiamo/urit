package urit

import (
	"errors"
	"github.com/go-andiamo/splitter"
	"net/http"
	"strings"
)

type Template interface {
	// PathFrom generates a path from the template given the specified path vars
	PathFrom(vars PathVars) (string, error)
	// Matches checks whether the specified path matches the template -
	// and if a successful match, returns the extracted path vars
	Matches(path string, options ...interface{}) (PathVars, bool)
	// MatchesRequest checks whether the specified request matches the template -
	// and if a successful match, returns the extracted path vars
	MatchesRequest(req *http.Request, options ...interface{}) (PathVars, bool)
	// Sub generates a new template with added sub-path
	Sub(path string, options ...interface{}) (Template, error)
	// ResolveTo generates a new template, filling in any known path vars from the supplied vars
	ResolveTo(vars PathVars) (Template, error)
	// OriginalTemplate returns the original (or generated) path template string
	OriginalTemplate() string
}

type template struct {
	originalTemplate string
	pathParts        []pathPart
	positionalVars   []pathPart
	namedVars        map[string][]pathPart
	posOnlyCount     int
	fixedMatchOpts   fixedMatchOptions
	varMatchOpts     varMatchOptions
	pathSplitOpts    []splitter.Option
}

// NewTemplate creates a new URI template from the path provided
//
// returns an error if the path cannot be parsed into a template
func NewTemplate(path string, options ...interface{}) (Template, error) {
	fs, vs, so := separateOptions(options)
	return (&template{
		originalTemplate: slashPrefix(path),
		pathParts:        make([]pathPart, 0),
		positionalVars:   make([]pathPart, 0),
		namedVars:        map[string][]pathPart{},
		posOnlyCount:     0,
		fixedMatchOpts:   fs,
		varMatchOpts:     vs,
		pathSplitOpts:    so,
	}).parse()
}

// MustCreateTemplate is the same as NewTemplate, except that it panics on error
func MustCreateTemplate(path string, options ...interface{}) Template {
	if t, err := NewTemplate(path, options...); err != nil {
		panic(err)
	} else {
		return t
	}
}

func separateOptions(options []interface{}) (fixedMatchOptions, varMatchOptions, []splitter.Option) {
	seenFixed := map[FixedMatchOption]bool{}
	seenVar := map[VarMatchOption]bool{}
	fixeds := make(fixedMatchOptions, 0)
	vars := make(varMatchOptions, 0)
	splitOps := make([]splitter.Option, 0)
	for _, intf := range options {
		if f, ok := intf.(FixedMatchOption); ok && !seenFixed[f] {
			fixeds = append(fixeds, f)
			seenFixed[f] = true
		} else if v, ok := intf.(VarMatchOption); ok && !seenVar[v] {
			vars = append(vars, v)
			seenVar[v] = true
		} else if s, ok := intf.(splitter.Option); ok {
			splitOps = append(splitOps, s)
		}
	}
	return fixeds, vars, splitOps
}

func (t *template) mergeOptions(options []interface{}) (fixedMatchOptions, varMatchOptions) {
	if len(options) == 0 {
		return t.fixedMatchOpts, t.varMatchOpts
	} else if len(t.fixedMatchOpts) == 0 && len(t.varMatchOpts) == 0 {
		fs, vs, _ := separateOptions(options)
		return fs, vs
	}
	fixed := make(fixedMatchOptions, 0)
	seenFixed := map[FixedMatchOption]bool{}
	vars := make(varMatchOptions, 0)
	seenVars := map[VarMatchOption]bool{}
	for _, f := range t.fixedMatchOpts {
		seenFixed[f] = true
		fixed = append(fixed, f)
	}
	for _, v := range t.varMatchOpts {
		seenVars[v] = true
		vars = append(vars, v)
	}
	for _, o := range options {
		if f, ok := o.(FixedMatchOption); ok && !seenFixed[f] {
			seenFixed[f] = true
			fixed = append(fixed, f)
		} else if v, ok := o.(VarMatchOption); ok && !seenVars[v] {
			seenVars[v] = true
			vars = append(vars, v)
		}
	}
	return fixed, vars
}

var uriSplitter = splitter.MustCreateSplitter('/',
	splitter.MustMakeEscapable(splitter.Parenthesis, '\\'),
	splitter.MustMakeEscapable(splitter.CurlyBrackets, '\\'),
	splitter.MustMakeEscapable(splitter.SquareBrackets, '\\'),
	splitter.DoubleQuotesBackSlashEscaped, splitter.SingleQuotesBackSlashEscaped).
	AddDefaultOptions(splitter.IgnoreEmptyOuters, splitter.NotEmptyInnersMsg("path parts cannot be empty"))

type partCapture struct {
	template *template
}

func (c *partCapture) Apply(s string, pos int, totalLen int, captured int, skipped int, isLast bool, subParts ...splitter.SubPart) (string, bool, error) {
	if pt, err := c.template.newUriPathPart(s, pos, subParts); err != nil {
		return "", false, err
	} else {
		c.template.pathParts = append(c.template.pathParts, pt)
	}
	return s, true, nil
}

func (t *template) clone() *template {
	result := &template{
		originalTemplate: t.originalTemplate,
		pathParts:        make([]pathPart, 0, len(t.pathParts)),
		positionalVars:   make([]pathPart, 0, len(t.positionalVars)),
		namedVars:        map[string][]pathPart{},
	}
	result.pathParts = append(result.pathParts, t.pathParts...)
	result.positionalVars = append(result.positionalVars, t.positionalVars...)
	for k, v := range t.namedVars {
		result.namedVars[k] = v
	}
	return result
}

func (t *template) parse() (Template, error) {
	if strings.Trim(t.originalTemplate, " ") == "" {
		return nil, newTemplateParseError("template empty", 0, nil)
	}
	splitOps := append(t.pathSplitOpts, &partCapture{template: t})
	_, err := uriSplitter.Split(t.originalTemplate, splitOps...)
	if t.posOnlyCount > 0 && len(t.namedVars) > 0 {
		return nil, newTemplateParseError("template cannot contain both positional and named path variables", 0, nil)
	}
	if err != nil {
		if terr := errors.Unwrap(err); terr != nil {
			if _, ok := terr.(TemplateParseError); ok {
				err = terr
			}
		}
	}
	return t, err
}

func (t *template) newUriPathPart(pt string, pos int, subParts []splitter.SubPart) (pathPart, error) {
	if len(subParts) == 1 && subParts[0].Type() == splitter.Fixed {
		if strings.HasPrefix(pt, "?") || strings.HasPrefix(pt, ":") {
			varPart := pathPart{
				fixed: false,
				name:  pt[1:],
			}
			t.addVar(varPart)
			return varPart, nil
		} else {
			return pathPart{
				fixed:      true,
				fixedValue: pt,
			}, nil
		}
	}
	return t.newVarPathPart(pt, pos, subParts)
}

func (t *template) addVar(pt pathPart) {
	if !pt.fixed {
		if pt.name != "" {
			t.namedVars[pt.name] = append(t.namedVars[pt.name], pt)
		} else {
			t.posOnlyCount++
		}
		t.positionalVars = append(t.positionalVars, pt)
	}
}

func (t *template) newVarPathPart(pt string, pos int, subParts []splitter.SubPart) (pathPart, error) {
	result := pathPart{
		fixed:    false,
		subParts: make([]pathPart, 0),
	}
	anyVarParts := false
	for _, sp := range subParts {
		if sp.Type() == splitter.Brackets && sp.StartRune() == '{' {
			anyVarParts = true
			str := sp.String()
			addPart := pathPart{
				fixed: false,
			}
			if err := addPart.setName(str[1:len(str)-1], sp.StartPos()); err != nil {
				return result, err
			}
			t.addVar(addPart)
			result.subParts = append(result.subParts, addPart)
		} else if sp.Type() == splitter.Quotes {
			result.subParts = append(result.subParts, pathPart{
				fixed:      true,
				fixedValue: sp.UnEscaped(),
			})
		} else {
			result.subParts = append(result.subParts, pathPart{
				fixed:      true,
				fixedValue: sp.String(),
			})
		}
	}
	if len(result.subParts) == 1 {
		return result.subParts[0], nil
	} else if !anyVarParts {
		var sb strings.Builder
		for _, s := range result.subParts {
			sb.WriteString(s.fixedValue)
		}
		return pathPart{
			fixed:      true,
			fixedValue: sb.String(),
		}, nil
	}
	return result, nil
}

func (t *template) PathFrom(vars PathVars) (string, error) {
	var pb strings.Builder
	tracker := &positionsTracker{
		vars:           vars,
		positional:     t.posOnlyCount > 0,
		position:       0,
		namedPositions: map[string]int{},
	}
	for _, pt := range t.pathParts {
		if str, err := pt.pathFrom(tracker); err == nil {
			pb.WriteString(str)
		} else {
			return "", err
		}
	}
	return pb.String(), nil
}

var matchPathSplitter = splitter.MustCreateSplitter('/').
	AddDefaultOptions(splitter.IgnoreEmptyFirst, splitter.IgnoreEmptyLast, splitter.NotEmptyInners)

func (t *template) Matches(path string, options ...interface{}) (PathVars, bool) {
	pts, err := matchPathSplitter.Split(path)
	if err != nil || len(pts) != len(t.pathParts) {
		return nil, false
	}
	result := newPathVars()
	fixedOpts, varOpts := t.mergeOptions(options)
	ok := true
	for i, pt := range t.pathParts {
		ok = pt.match(pts[i], i, result, fixedOpts, varOpts)
		if !ok {
			break
		}
	}
	return result, ok
}

func (t *template) MatchesRequest(req *http.Request, options ...interface{}) (PathVars, bool) {
	return t.Matches(req.URL.Path, options...)
}

func (t *template) Sub(path string, options ...interface{}) (Template, error) {
	add, err := NewTemplate(path, options...)
	if err != nil {
		return nil, err
	}
	ra, _ := add.(*template)
	if (ra.posOnlyCount > 0 && len(t.namedVars) > 0) || (t.posOnlyCount > 0 && len(ra.namedVars) > 0) {
		return nil, newTemplateParseError("template cannot contain both positional and named path variables", 0, nil)
	}
	result := t.clone()
	if strings.HasSuffix(result.originalTemplate, "/") {
		result.originalTemplate = result.originalTemplate[:len(result.originalTemplate)-1] + ra.originalTemplate
	} else {
		result.originalTemplate = result.originalTemplate + ra.originalTemplate
	}
	for _, pt := range ra.pathParts {
		result.pathParts = append(result.pathParts, pt)
	}
	for _, argPt := range ra.positionalVars {
		result.addVar(argPt)
	}
	return result, nil
}

func (t *template) ResolveTo(vars PathVars) (Template, error) {
	tracker := &positionsTracker{
		vars:           vars,
		positional:     t.posOnlyCount > 0,
		position:       0,
		namedPositions: map[string]int{},
	}
	result := &template{
		pathParts:      make([]pathPart, 0, len(t.pathParts)),
		positionalVars: make([]pathPart, 0, len(t.positionalVars)),
		namedVars:      map[string][]pathPart{},
		posOnlyCount:   0,
	}
	var orgBuilder strings.Builder
	for _, pt := range t.pathParts {
		if pt.fixed {
			orgBuilder.WriteString(`/` + pt.fixedValue)
			result.pathParts = append(result.pathParts, pt)
		} else if len(pt.subParts) == 0 {
			if str, err := tracker.getVar(&pt); err == nil {
				orgBuilder.WriteString(`/` + str)
				result.pathParts = append(result.pathParts, pathPart{
					fixed:      true,
					fixedValue: str,
				})
			} else {
				result.pathParts = append(result.pathParts, pt)
				if pt.name == "" {
					result.posOnlyCount++
					result.positionalVars = append(result.positionalVars, pt)
					orgBuilder.WriteString(`/?`)
				} else {
					orgBuilder.WriteString(`/{` + pt.name)
					result.namedVars[pt.name] = append(result.namedVars[pt.name], pt)
					if pt.orgRegexp != "" {
						orgBuilder.WriteString(`:` + pt.orgRegexp)
					}
					orgBuilder.WriteString(`}`)
				}
			}
		} else {
			np := pathPart{
				fixed:    false,
				subParts: make([]pathPart, 0, len(pt.subParts)),
			}
			resolvedCount := 0
			for _, sp := range pt.subParts {
				if sp.fixed {
					resolvedCount++
					np.subParts = append(np.subParts, sp)
				} else if str, err := tracker.getVar(&sp); err == nil {
					resolvedCount++
					np.subParts = append(np.subParts, pathPart{
						fixed:      true,
						fixedValue: str,
					})
				} else {
					np.subParts = append(np.subParts, sp)
					result.namedVars[sp.name] = append(result.namedVars[sp.name], sp)
				}
			}
			orgBuilder.WriteString(`/`)
			if resolvedCount == len(pt.subParts) {
				fxnp := pathPart{
					fixed:      true,
					fixedValue: "",
				}
				for _, sp := range np.subParts {
					orgBuilder.WriteString(sp.fixedValue)
					fxnp.fixedValue += sp.fixedValue
				}
				np = fxnp
			} else {
				for _, sp := range np.subParts {
					if sp.fixed {
						orgBuilder.WriteString(sp.fixedValue)
					} else {
						orgBuilder.WriteString(`{` + sp.name)
						if sp.orgRegexp != "" {
							orgBuilder.WriteString(`:` + sp.orgRegexp)
						}
						orgBuilder.WriteString(`}`)
					}
				}
			}
			result.pathParts = append(result.pathParts, np)
		}
	}
	result.originalTemplate = orgBuilder.String()
	return result, nil
}

func (t *template) OriginalTemplate() string {
	return t.originalTemplate
}

func slashPrefix(s string) string {
	if strings.Trim(s, " ") == "" {
		return s
	} else if strings.HasPrefix(s, "/") {
		return s
	}
	return "/" + s
}