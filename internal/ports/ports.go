package ports

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Range struct {
	Start int
	End   int
}

type Set struct {
	ranges []Range
}

func Parse(spec string) (Set, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return Set{ranges: []Range{{Start: 1, End: 65535}}}, nil
	}
	parts := strings.Split(spec, ",")
	out := make([]Range, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "-") {
			bounds := strings.SplitN(p, "-", 2)
			if len(bounds) != 2 {
				return Set{}, fmt.Errorf("invalid port range: %q", p)
			}
			a, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return Set{}, fmt.Errorf("invalid port %q: %w", bounds[0], err)
			}
			b, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return Set{}, fmt.Errorf("invalid port %q: %w", bounds[1], err)
			}
			if a < 1 || b < 1 || a > 65535 || b > 65535 || a > b {
				return Set{}, fmt.Errorf("invalid port range: %q", p)
			}
			out = append(out, Range{Start: a, End: b})
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return Set{}, fmt.Errorf("invalid port %q: %w", p, err)
		}
		if n < 1 || n > 65535 {
			return Set{}, fmt.Errorf("invalid port %q", p)
		}
		out = append(out, Range{Start: n, End: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start < out[j].Start })
	return Set{ranges: merge(out)}, nil
}

func merge(in []Range) []Range {
	if len(in) == 0 {
		return nil
	}
	merged := []Range{in[0]}
	for _, r := range in[1:] {
		last := &merged[len(merged)-1]
		if r.Start <= last.End+1 {
			if r.End > last.End {
				last.End = r.End
			}
			continue
		}
		merged = append(merged, r)
	}
	return merged
}

func (s Set) Contains(port int) bool {
	for _, r := range s.ranges {
		if port >= r.Start && port <= r.End {
			return true
		}
	}
	return false
}
