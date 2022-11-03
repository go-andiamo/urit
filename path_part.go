package urit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type pathPart struct {
	fixed         bool
	fixedValue    string
	subParts      []pathPart
	regexp        *regexp.Regexp
	orgRegexp     string
	allRegexp     *regexp.Regexp
	allRegexpIdxs map[int]int
	name          string
}

func (pt *pathPart) setName(name string, pos int) error {
	if cAt := strings.IndexByte(name, ':'); cAt != -1 {
		pt.name = strings.Trim(name[:cAt], " ")
		if pt.name == "" {
			return newTemplateParseError("path var name cannot be empty", pos, nil)
		}
		pt.orgRegexp = strings.Trim(name[cAt+1:], " ")
		if pt.orgRegexp != "" {
			rxBit := addRegexHeadAndTail(pt.orgRegexp)
			if rx, err := regexp.Compile(rxBit); err == nil {
				pt.regexp = rx
			} else {
				return newTemplateParseError("path var regexp problem", pos+cAt, err)
			}
		}
	} else {
		pt.name = strings.Trim(name, " ")
	}
	if pt.name == "" {
		return newTemplateParseError("path var name cannot be empty", pos, nil)
	}
	return nil
}

func (pt *pathPart) addFound(vars PathVars, val string) {
	if pt.name != "" {
		vars.AddNamedValue(pt.name, val)
	} else {
		vars.AddPositionalValue(val)
	}
}

func (pt *pathPart) match(s string, pathPos int, vars PathVars, fOpts fixedMatchOptions, vOpts varMatchOptions) bool {
	if pt.fixed {
		ok := pt.fixedValue == s
		if len(fOpts) > 0 {
			ok = fOpts.check(s, pt.fixedValue, pathPos, vars)
		}
		return ok
	} else if len(pt.subParts) == 0 {
		ok := pt.regexp == nil || pt.regexp.MatchString(s)
		if len(vOpts) > 0 {
			if rs, vok, applicable := vOpts.check(s, vars.Len(), pt.name, pt.regexp, pt.orgRegexp, pathPos, vars); applicable {
				s = rs
				ok = vok
			}
		}
		if ok {
			pt.addFound(vars, s)
			return true
		}
	} else {
		return pt.multiMatch(s, pathPos, vars, vOpts)
	}
	return false
}

func (pt *pathPart) multiMatch(s string, pathPos int, vars PathVars, vOpts varMatchOptions) bool {
	orx := pt.overallRegexp()
	sms := orx.FindStringSubmatch(s)
	if len(sms) > 0 {
		for i, sp := range pt.subParts {
			if !sp.fixed {
				str := sms[pt.allRegexpIdxs[i]]
				ok := true
				if len(vOpts) > 0 {
					if rs, vok, applicable := vOpts.check(str, vars.Len(), sp.name, sp.regexp, sp.orgRegexp, pathPos, vars); applicable {
						str = rs
						ok = vok
					}
				}
				if ok {
					sp.addFound(vars, str)
				}
			}
		}
		return true
	}
	return false
}

func (pt *pathPart) overallRegexp() *regexp.Regexp {
	if pt.allRegexp == nil {
		var rxb strings.Builder
		for i, sp := range pt.subParts {
			if sp.fixed {
				rxb.WriteString(`(\Q` + sp.fixedValue + `\E)`)
			} else if sp.orgRegexp != "" {
				rxb.WriteString(`(?P<vsp` + fmt.Sprintf("%d", i) + `>` + stripRegexHeadAndTail(sp.orgRegexp) + `)`)
			} else {
				rxb.WriteString(`(?P<vsp` + fmt.Sprintf("%d", i) + `>.*)`)
			}
		}
		if rx, err := regexp.Compile(addRegexHeadAndTail(rxb.String())); err == nil {
			pt.allRegexp = rx
			pt.allRegexpIdxs = map[int]int{}
			for i, nm := range rx.SubexpNames() {
				if i > 0 && nm != "" && strings.HasPrefix(nm, "vsp") {
					nmi, _ := strconv.Atoi(nm[3:])
					pt.allRegexpIdxs[nmi] = i
				}
			}
		}
	}
	return pt.allRegexp
}

func (pt *pathPart) pathFrom(tracker *positionsTracker) (string, error) {
	if pt.fixed {
		return `/` + pt.fixedValue, nil
	} else if len(pt.subParts) == 0 {
		if str, err := tracker.getVar(pt); err == nil {
			return `/` + str, nil
		} else {
			return "", err
		}
	}
	var pb strings.Builder
	for _, sp := range pt.subParts {
		if sp.fixed {
			pb.WriteString(sp.fixedValue)
		} else if str, err := tracker.getVar(&sp); err == nil {
			pb.WriteString(str)
		} else {
			return "", err
		}
	}
	return `/` + pb.String(), nil
}

type positionsTracker struct {
	vars           PathVars
	positional     bool
	position       int
	namedPositions map[string]int
}

func (t *positionsTracker) getVar(pt *pathPart) (string, error) {
	if t.positional {
		if str, ok := t.vars.GetPositional(t.position); ok {
			t.position++
			return str, nil
		}
		return "", fmt.Errorf("no var for position %d", t.position+1)
	} else {
		np := t.namedPositions[pt.name]
		if str, ok := t.vars.GetNamed(pt.name, np); ok {
			t.namedPositions[pt.name] = np + 1
			return str, nil
		} else if np == 0 {
			return "", fmt.Errorf("no var for '%s'", pt.name)
		}
		return "", fmt.Errorf("no var for '%s' (position %d)", pt.name, np+1)
	}
}

func addRegexHeadAndTail(rx string) string {
	head := ""
	tail := ""
	if !strings.HasPrefix(rx, "^") {
		head = "^"
	}
	if !strings.HasSuffix(rx, "$") {
		tail = "$"
	}
	return head + rx + tail
}

func stripRegexHeadAndTail(rx string) string {
	if strings.HasPrefix(rx, "^") {
		rx = rx[1:]
	}
	if strings.HasSuffix(rx, "$") {
		rx = rx[:len(rx)-1]
	}
	return rx
}
