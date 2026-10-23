package parser

import (
	"regexp"
	"strings"
)

// IgnoreMatcher biên dịch danh sách pattern ignore một lần và cung cấp
// tra cứu nhanh cho từng package ID. Thay vì quét lại toàn bộ pattern
// cho mỗi package (O(N×M)), matcher này tách thành các loại match O(1)
// hoặc O(prefix) giúp giảm đáng dù số lần lặp khi scan project lớn.
type IgnoreMatcher struct {
	exact  map[string]struct{}
	prefix []string
	glob   []globEntry
}

// globEntry gắn pattern glob với regex đã compile để tránh gọi path.Match
// lặp đi lặp lại cho mỗi package.
type globEntry struct {
	pattern string
	regex   *regexp.Regexp
}

// NewIgnoreMatcher biên dịch danh sách pattern ignore từ config/CLI.
//
// Hỗ trợ các dạng:
//   - "github.com/foo/bar"          -> exact match
//   - "github.com/foo/bar/..."      -> prefix match (bar và mọi sub-package)
//   - "github.com/foo/bar/*"        -> exact match thư mục cha + prefix sub-package
//   - "github.com/foo/*bar*"        -> glob match (compile thành regex)
//   - "github.com/foo/bar"          -> exact match (chỉ khớp chính xác gói này,
//     không khớp sub-package, tránh false-positive)
func NewIgnoreMatcher(patterns []string) (*IgnoreMatcher, error) {
	m := &IgnoreMatcher{
		exact: make(map[string]struct{}),
	}

	for _, pat := range patterns {
		pat = strings.TrimSpace(pat)
		if pat == "" {
			continue
		}

		switch {
		case strings.HasSuffix(pat, "/*"):
			prefix := strings.TrimSuffix(strings.TrimSuffix(pat, "/*"), "/")
			// "foo/bar/*" khớp chính xác "foo/bar" và mọi gói bên dưới "foo/bar/"
			m.exact[prefix] = struct{}{}
			m.prefix = append(m.prefix, prefix)

		case strings.HasSuffix(pat, "..."):
			prefix := strings.TrimSuffix(strings.TrimSuffix(pat, "..."), "/")
			m.prefix = append(m.prefix, prefix)

		case strings.Contains(pat, "*"):
			re, err := globToRegex(pat)
			if err != nil {
				return nil, err
			}
			m.glob = append(m.glob, globEntry{pattern: pat, regex: re})

		default:
			// Pattern thuần túy: exact match O(1), không gây false-positive
			// (ví dụ "foo" sẽ KHÔNG khớp "foobar").
			m.exact[pat] = struct{}{}
		}
	}

	return m, nil
}

// globToRegex chuyển một pattern glob (path.Match style) thành regexp.
// Chỉ hỗ trợ ký tự wildcard '*' (không hỗ trợ '?' để tránh nhầm lẫn).
func globToRegex(pat string) (*regexp.Regexp, error) {
	var sb strings.Builder
	sb.WriteString("^")
	for _, ch := range pat {
		switch ch {
		case '*':
			sb.WriteString("[^/]*")
		case '.', '+', '(', ')', '[', ']', '{', '}', '|', '\\', '^', '$':
			sb.WriteString("\\")
			sb.WriteRune(ch)
		default:
			sb.WriteRune(ch)
		}
	}
	sb.WriteString("$")
	return regexp.Compile(sb.String())
}

// ShouldIgnore trả về true nếu package ID bị bỏ qua theo matcher.
func (m *IgnoreMatcher) ShouldIgnore(pkgID string) bool {
	if _, ok := m.exact[pkgID]; ok {
		return true
	}
	for _, prefix := range m.prefix {
		if pkgID == prefix || strings.HasPrefix(pkgID, prefix+"/") {
			return true
		}
	}
	for _, g := range m.glob {
		if g.regex.MatchString(pkgID) {
			return true
		}
	}
	return false
}
