package matching

import (
	"strings"
)

func JaroWinkler(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	r1 := []rune(strings.ToUpper(strings.TrimSpace(s1)))
	r2 := []rune(strings.ToUpper(strings.TrimSpace(s2)))

	len1 := len(r1)
	len2 := len(r2)

	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	matchDist := max(len1, len2)/2 - 1
	if matchDist < 0 {
		matchDist = 0
	}

	m1 := make([]bool, len1)
	m2 := make([]bool, len2)

	matches := 0
	for i := 0; i < len1; i++ {
		start := i - matchDist
		if start < 0 {
			start = 0
		}
		end := i + matchDist + 1
		if end > len2 {
			end = len2
		}
		for j := start; j < end; j++ {
			if m1[i] || m2[j] {
				continue
			}
			if r1[i] != r2[j] {
				continue
			}
			m1[i] = true
			m2[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	transpositions := 0
	k := 0
	for i := 0; i < len1; i++ {
		if !m1[i] {
			continue
		}
		for j := k; j < len2; j++ {
			if !m2[j] {
				continue
			}
			if r1[i] != r2[j] {
				transpositions++
			}
			k = j + 1
			break
		}
	}

	jaro := (float64(matches)/float64(len1) +
		float64(matches)/float64(len2) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0

	prefix := 0
	limit := len1
	if len2 < limit {
		limit = len2
	}
	if limit > 4 {
		limit = 4
	}
	for i := 0; i < limit; i++ {
		if r1[i] == r2[i] {
			prefix++
		} else {
			break
		}
	}

	return jaro + float64(prefix)*0.1*(1.0-jaro)
}

func Levenshtein(s1, s2 string) int {
	r1 := []rune(strings.ToUpper(strings.TrimSpace(s1)))
	r2 := []rune(strings.ToUpper(strings.TrimSpace(s2)))

	if len(r1) == 0 {
		return len(r2)
	}
	if len(r2) == 0 {
		return len(r1)
	}

	prev := make([]int, len(r2)+1)
	curr := make([]int, len(r2)+1)

	for j := 0; j <= len(r2); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(r1); i++ {
		curr[0] = i
		for j := 1; j <= len(r2); j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min(del, min(ins, sub))
		}
		prev, curr = curr, prev
	}

	return prev[len(r2)]
}

func NormalizedLevenshtein(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	r1 := []rune(strings.ToUpper(strings.TrimSpace(s1)))
	r2 := []rune(strings.ToUpper(strings.TrimSpace(s2)))
	if len(r1) == 0 || len(r2) == 0 {
		return 0.0
	}
	levDist := Levenshtein(s1, s2)
	maxLen := len(r1)
	if len(r2) > maxLen {
		maxLen = len(r2)
	}
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(levDist)/float64(maxLen)
}

func Soundex(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return ""
	}

	code := map[byte]byte{
		'B': '1', 'F': '1', 'P': '1', 'V': '1',
		'C': '2', 'G': '2', 'J': '2', 'K': '2', 'Q': '2', 'S': '2', 'X': '2', 'Z': '2',
		'D': '3', 'T': '3',
		'L': '4',
		'M': '5', 'N': '5',
		'R': '6',
	}

	result := []byte{s[0]}
	lastCode := code[s[0]]

	for i := 1; i < len(s) && len(result) < 4; i++ {
		ch := s[i]
		c, ok := code[ch]
		if !ok {
			lastCode = 0
			continue
		}
		if c != lastCode {
			result = append(result, c)
			lastCode = c
		}
	}

	for len(result) < 4 {
		result = append(result, '0')
	}

	return string(result[:4])
}

func Metaphone(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(s))

	i := 0
	n := len(s)

	if n > 1 {
		prefix := s[:2]
		if prefix == "KN" || prefix == "GN" || prefix == "PN" || prefix == "WR" || prefix == "PS" {
			i = 1
		}
	}

	for i < n {
		ch := s[i]

		if ch < 'A' || ch > 'Z' {
			i++
			continue
		}

		if strings.ContainsRune("AEIOU", rune(ch)) {
			if i == 0 {
				result.WriteByte(ch)
			}
			i++
			continue
		}

		switch ch {
		case 'B':
			if i == 0 || i != n-1 || s[i-1] != 'M' {
				result.WriteByte('B')
			}
		case 'C':
			lookahead := byte(0)
			if i+1 < n {
				lookahead = s[i+1]
			}
			if lookahead == 'H' {
				result.WriteByte('X')
				i++
			} else if lookahead == 'I' && i+2 < n && s[i+2] == 'A' {
				result.WriteByte('X')
			} else if lookahead == 'I' || lookahead == 'E' || lookahead == 'Y' {
				result.WriteByte('S')
			} else {
				result.WriteByte('K')
			}
		case 'D':
			if i+2 < n && s[i+1] == 'G' && (s[i+2] == 'E' || s[i+2] == 'Y') {
				result.WriteByte('J')
				i += 2
				continue
			}
			result.WriteByte('T')
		case 'F':
			result.WriteByte('F')
		case 'G':
			lookahead := byte(0)
			if i+1 < n {
				lookahead = s[i+1]
			}
			if i+1 < n && lookahead == 'H' {
				// silent GH
			} else if i+2 < n && lookahead == 'N' {
				// silent GN
			} else if lookahead == 'E' || lookahead == 'Y' {
				result.WriteByte('J')
			} else {
				result.WriteByte('K')
			}
		case 'H':
			if i > 0 && i < n-1 && (s[i-1] == 'A' || s[i-1] == 'E' || s[i-1] == 'I' || s[i-1] == 'O' || s[i-1] == 'U') {
				// silent H after vowel
			} else if i < n-1 && strings.ContainsRune("AEIOU", rune(s[i+1])) {
				result.WriteByte('H')
			}
		case 'J':
			result.WriteByte('J')
		case 'K':
			if i == 0 || s[i-1] != 'C' {
				result.WriteByte('K')
			}
		case 'L':
			result.WriteByte('L')
		case 'M':
			result.WriteByte('M')
		case 'N':
			result.WriteByte('N')
		case 'P':
			if i+1 < n && s[i+1] == 'H' {
				result.WriteByte('F')
				i++
			} else {
				result.WriteByte('P')
			}
		case 'Q':
			result.WriteByte('K')
		case 'R':
			result.WriteByte('R')
		case 'S':
			lookahead := byte(0)
			if i+1 < n {
				lookahead = s[i+1]
			}
			if lookahead == 'H' {
				result.WriteByte('X')
				i++
			} else if lookahead == 'I' && i+2 < n && s[i+2] == 'O' {
				result.WriteByte('X')
			} else {
				result.WriteByte('S')
			}
		case 'T':
			if i+2 < n && s[i+1] == 'I' && s[i+2] == 'O' {
				result.WriteByte('X')
			} else if i+1 < n && s[i+1] == 'H' {
				result.WriteByte('0')
				i++
			} else {
				result.WriteByte('T')
			}
		case 'V':
			result.WriteByte('F')
		case 'W':
			if i+1 < n && strings.ContainsRune("AEIOU", rune(s[i+1])) {
				result.WriteByte('W')
			}
		case 'X':
			result.WriteByte('K')
			result.WriteByte('S')
		case 'Y':
			if i+1 < n && !strings.ContainsRune("AEIOU", rune(s[i+1])) {
				result.WriteByte('Y')
			}
		case 'Z':
			result.WriteByte('S')
		}

		i++
	}

	return result.String()
}

func NameSimilarity(first1, last1, first2, last2 string) float64 {
	f1 := strings.ToUpper(strings.TrimSpace(first1))
	l1 := strings.ToUpper(strings.TrimSpace(last1))
	f2 := strings.ToUpper(strings.TrimSpace(first2))
	l2 := strings.ToUpper(strings.TrimSpace(last2))

	fnSim := JaroWinkler(f1, f2)
	fnLev := NormalizedLevenshtein(f1, f2)
	lnSim := JaroWinkler(l1, l2)
	lnLev := NormalizedLevenshtein(l1, l2)

	fnScore := 0.6*fnSim + 0.4*fnLev
	lnScore := 0.6*lnSim + 0.4*lnLev

	return 0.4*fnScore + 0.6*lnScore
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
