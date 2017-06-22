package lookup

import (
	"regexp"
)

// ArgsLooker looks up keys in []string, like the one in os.Args.
type ArgsLooker struct {
	args []string
	rex  *regexp.Regexp

	extraArgs []string
	data      Map
}

// NewArgs returns a Looker that can extract data from program arguments (e.g, os.Args).
// Suggestions for prefix: "-", "--env-" or even "".
// Valid args (<prefix><NAME>=<value>) are processed by LookupKey and the rest is available with ExtraArgs.
func NewArgs(prefix string, args []string) *ArgsLooker {
	l := ArgsLooker{
		rex:  regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `([^=]*)(?:(=)(.*))?$`),
		args: make([]string, len(args)),
	}
	copy(l.args, args)
	return &l
}

// ExtraArgs returns args that are not formatted for ArgsLooker. Generally your program should process them.
// It is empty before the first call to LookupKey.
func (l *ArgsLooker) ExtraArgs() []string {
	return l.extraArgs
}

// LookupKey processes provided args (1st call only) and looks up the value of k.
func (l *ArgsLooker) LookupKey(k string) (string, bool, error) {
	if l.data == nil {
		l.data = make(Map)

		for _, arg := range l.args {
			res := l.rex.FindStringSubmatch(arg)
			if len(res) < 4 {
				l.extraArgs = append(l.extraArgs, arg)
				continue
			}

			val := res[3]
			if res[2] == "" {
				// default for syntax "-KEY" is 1, which Lookup can save into an int, bool, etc.
				val = "1"
			}
			l.data[res[1]] = val
		}
	}
	return l.data.LookupKey(k)
}
