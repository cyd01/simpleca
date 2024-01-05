package tools

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Remove duplicate entry in an aray of string
func RemoveDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

// Test is string is a int
func IsInt(s string) bool {
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return true
	}
	return false
}

// Convert a string into a int64
func Atoi64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Convert a string into a float64
func Atof64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// Convert a int64 into a string
func I64toa(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Convert a float64 into a string
func F64toa(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// Test if s string match p regex pattern
func Match(p string, s string) bool {
	match, err := regexp.MatchString(p, s)
	if err != nil {
		return false
	}
	return match
}

// Test if s string match p pattern with joker (*, ?)
func IsMatch(s string, p string) bool {
	runeInput := []rune(s)
	runePattern := []rune(p)
	lenInput := len(runeInput)
	lenPattern := len(runePattern)
	isMatchingMatrix := make([][]bool, lenInput+1)
	for i := range isMatchingMatrix {
		isMatchingMatrix[i] = make([]bool, lenPattern+1)
	}
	isMatchingMatrix[0][0] = true
	for i := 1; i < lenInput; i++ {
		isMatchingMatrix[i][0] = false
	}
	if lenPattern > 0 {
		if runePattern[0] == '*' {
			isMatchingMatrix[0][1] = true
		}
	}
	for j := 2; j <= lenPattern; j++ {
		if runePattern[j-1] == '*' {
			isMatchingMatrix[0][j] = isMatchingMatrix[0][j-1]
		}

	}
	for i := 1; i <= lenInput; i++ {
		for j := 1; j <= lenPattern; j++ {

			if runePattern[j-1] == '*' {
				isMatchingMatrix[i][j] = isMatchingMatrix[i-1][j] || isMatchingMatrix[i][j-1]
			}

			if runePattern[j-1] == '?' || runeInput[i-1] == runePattern[j-1] {
				isMatchingMatrix[i][j] = isMatchingMatrix[i-1][j-1]
			}
		}
	}
	return isMatchingMatrix[lenInput][lenPattern]
}

// Replace p regex pattern with r string in s string
func Replace(p string, r string, s string) string {
	var re = regexp.MustCompile(p)
	return re.ReplaceAllString(s, r)
}

// Replace \r\n or \n with r string in s string
func Replacecr(r string, s string) string {
	var re = regexp.MustCompile("(\r)?\n")
	return re.ReplaceAllString(s, r)
}

// exists returns whether the given file or directory exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// existsdir returns whether the given path exists and is a directory
func Existsdir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return true, nil
		} else {
			return true, nil
		}
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// existsfile return if the given path exists and is a regular file
func Existsfile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !fi.Mode().IsRegular() {
		return false
	}
	return true
}

// genuuid returns a valid uniq uuid under a string
func Genuuid() string {
	return uuid.New().String()
}

// Generate random string
func RandomString(pattern string, n int) string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune(pattern)
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// Test host for opened ports
func Raw_connect(host string, ports []string) bool {
	for _, port := range ports {
		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Connecting error:", err)
			return false
		}
		if conn != nil {
			defer conn.Close()
			fmt.Fprintln(os.Stderr, "Opened", net.JoinHostPort(host, port))
			return true
		}
	}
	return false
}

// Test if file exists
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Read first line
func FileFirstLine(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// Remove suffix from string
func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

// Add suffix if not exists
func PadSuffix(s, suffix string) string {
	if !strings.HasSuffix(s, suffix) {
		s = s + suffix
	}
	return s
}

// Return epoch in seconds
func Epoch() int64 {
	now := time.Now()
	secs := now.Unix()
	//nanos := now.UnixNano()
	//millis := nanos/1000000
	return secs
}

// Calculate max length of a multi-line string
func Maxlen(s string) int {
	l := 0
	v := strings.Split(s, "\n")
	for _, m := range v {
		if len(m) > l {
			l = len(m)
		}
	}
	return l
}

// Insert a char every nth characters of a string (useful to add \n for printing long one-line)
func InsertNth(s string, n int, c string) string {
	for i := n; i < len(s); i += (n + 1) {
		s = s[:i] + c + s[i:]
	}
	return s
}

// Parcours une chaine et ajoute des \n tous les n characteres si besoin (s'il n'y en a pas déjà)
func EnsureStringNotLong(s string, l int) string {
	st := ""
	n := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == "\n" {
			st = st + string(s[i])
			n = 0
		} else if n >= l {
			st = st + "\n" + string(s[i])
			n = 1
		} else {
			st = st + string(s[i])
			n++
		}
	}
	return st
}

/* Generic function to test is a value exists in aslice */
func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
