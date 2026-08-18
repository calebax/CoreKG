package utils

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

var (
	pdfPagesCountRegexp = regexp.MustCompile(`/Type\s*/Pages\b[\s\S]*?/Count\s+(\d+)`)
	pdfPageRegexp       = regexp.MustCompile(`/Type\s*/Page\b`)
	pdfTextArrayRegexp  = regexp.MustCompile(`\[(.*?)\]\s*TJ`)
	pdfTextShowRegexp   = regexp.MustCompile(`(\((?:\\.|[^\\()])*\)|<[0-9A-Fa-f\s]+>)\s*(?:Tj|\'|")`)
	pdfTextItemRegexp   = regexp.MustCompile(`\((?:\\.|[^\\()])*\)|<[0-9A-Fa-f\s]+>`)
	pdfBFCharRegexp     = regexp.MustCompile(`^\s*<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>\s*$`)
	pdfBFRangeRegexp    = regexp.MustCompile(`^\s*<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>\s*(<([0-9A-Fa-f]+)>|\[(.*)\])\s*$`)
	pdfHexTokenRegexp   = regexp.MustCompile(`<([0-9A-Fa-f]+)>`)
)

type PDFFileInfo struct {
	PageCount int `json:"pageCount"`
}

type pdfCMap struct {
	mappingByWidth map[int]map[string]string
	widths         []int
}

func GetPDFFileInfo(pathOrURL string) (*PDFFileInfo, error) {
	reader, closeFn, err := openPDFReader(pathOrURL)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	pageCount, err := getPDFPageCount(reader)
	if err != nil {
		return nil, err
	}

	return &PDFFileInfo{
		PageCount: pageCount,
	}, nil
}

func ReadPDFLinesRange(reader io.Reader, startLine, endLine int) ([]string, error) {
	text, err := extractPDFText(reader)
	if err != nil {
		return nil, err
	}
	return ReadLinesRange(strings.NewReader(text), startLine, endLine)
}

func getPDFPageCount(reader io.Reader) (int, error) {
	const maxPDFInspectSize = 64 << 20
	data, err := io.ReadAll(io.LimitReader(reader, maxPDFInspectSize+1))
	if err != nil {
		return 0, fmt.Errorf("read pdf content failed: %w", err)
	}
	if len(data) > maxPDFInspectSize {
		return 0, fmt.Errorf("pdf is too large to inspect for page count")
	}

	if matches := pdfPagesCountRegexp.FindAllSubmatch(data, -1); len(matches) > 0 {
		maxCount := 0
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			count, err := strconv.Atoi(string(match[1]))
			if err != nil {
				continue
			}
			if count > maxCount {
				maxCount = count
			}
		}
		if maxCount > 0 {
			return maxCount, nil
		}
	}

	pageCount := len(pdfPageRegexp.FindAll(data, -1))
	if pageCount > 0 {
		return pageCount, nil
	}

	return 0, fmt.Errorf("page count metadata not found")
}

func openPDFReader(pathOrURL string) (io.ReadCloser, func(), error) {
	if u, err := neturl.Parse(pathOrURL); err == nil && u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "http", "https":
			resp, err := http.Get(pathOrURL)
			if err != nil {
				return nil, nil, fmt.Errorf("http request failed: %v", err)
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				defer resp.Body.Close()
				return nil, nil, fmt.Errorf("http request failed: %s", resp.Status)
			}
			return resp.Body, func() { _ = resp.Body.Close() }, nil
		case "file":
			localPath := filepath.FromSlash(u.Path)
			file, err := os.Open(localPath)
			if err != nil {
				return nil, nil, fmt.Errorf("open local file failed: %v", err)
			}
			return file, func() { _ = file.Close() }, nil
		default:
			return nil, nil, fmt.Errorf("不支持的URL协议: %s", u.Scheme)
		}
	}

	file, err := os.Open(pathOrURL)
	if err != nil {
		return nil, nil, fmt.Errorf("open local file failed: %v", err)
	}
	return file, func() { _ = file.Close() }, nil
}

func extractPDFText(reader io.Reader) (string, error) {
	const maxPDFInspectSize = 64 << 20
	data, err := io.ReadAll(io.LimitReader(reader, maxPDFInspectSize+1))
	if err != nil {
		return "", fmt.Errorf("read pdf content failed: %w", err)
	}
	if len(data) > maxPDFInspectSize {
		return "", fmt.Errorf("pdf is too large to inspect")
	}

	streams := extractPDFStreams(data)
	cmap := newPDFCMap()
	for _, stream := range streams {
		cmap.merge(extractPDFCMap(stream))
	}

	fragments := make([]string, 0)
	for _, stream := range streams {
		fragments = append(fragments, extractTextFragments(stream, cmap)...)
	}
	if len(fragments) == 0 {
		fragments = extractTextFragments(data, cmap)
	}
	if len(fragments) == 0 {
		return "", fmt.Errorf("no readable text found in pdf")
	}

	return strings.Join(fragments, "\n"), nil
}

func extractPDFStreams(data []byte) [][]byte {
	streams := make([][]byte, 0)
	searchFrom := 0
	for {
		streamIdx := bytes.Index(data[searchFrom:], []byte("stream"))
		if streamIdx < 0 {
			break
		}
		streamIdx += searchFrom

		contentStart := streamIdx + len("stream")
		if contentStart < len(data) && data[contentStart] == '\r' {
			contentStart++
		}
		if contentStart < len(data) && data[contentStart] == '\n' {
			contentStart++
		}

		endIdx := bytes.Index(data[contentStart:], []byte("endstream"))
		if endIdx < 0 {
			break
		}
		endIdx += contentStart

		streamData := data[contentStart:endIdx]
		dictStart := streamIdx - 512
		if dictStart < 0 {
			dictStart = 0
		}
		dictData := data[dictStart:streamIdx]
		if bytes.Contains(dictData, []byte("/FlateDecode")) {
			if decoded, err := decodeFlateStream(streamData); err == nil && len(decoded) > 0 {
				streams = append(streams, decoded)
			}
		}
		streams = append(streams, streamData)
		searchFrom = endIdx + len("endstream")
	}
	return streams
}

func decodeFlateStream(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(bytes.TrimSpace(data)))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func extractTextFragments(data []byte, cmap *pdfCMap) []string {
	fragments := make([]string, 0)
	seen := make(map[string]struct{})

	for _, match := range pdfTextArrayRegexp.FindAllSubmatch(data, -1) {
		if len(match) < 2 {
			continue
		}
		parts := make([]string, 0)
		for _, item := range pdfTextItemRegexp.FindAll(match[1], -1) {
			if text := decodePDFTextToken(item, cmap); text != "" {
				parts = append(parts, text)
			}
		}
		if text := normalizePDFText(strings.Join(parts, "")); text != "" {
			if _, ok := seen[text]; !ok {
				seen[text] = struct{}{}
				fragments = append(fragments, text)
			}
		}
	}

	for _, match := range pdfTextShowRegexp.FindAllSubmatch(data, -1) {
		if len(match) < 2 {
			continue
		}
		if text := decodePDFTextToken(match[1], cmap); text != "" {
			if _, ok := seen[text]; !ok {
				seen[text] = struct{}{}
				fragments = append(fragments, text)
			}
		}
	}

	return fragments
}

func decodePDFTextToken(token []byte, cmap *pdfCMap) string {
	token = bytes.TrimSpace(token)
	if len(token) < 2 {
		return ""
	}

	switch token[0] {
	case '(':
		return normalizePDFText(decodePDFLiteralString(token[1:len(token)-1], cmap))
	case '<':
		return normalizePDFText(decodePDFHexString(token[1:len(token)-1], cmap))
	default:
		return ""
	}
}

func decodePDFLiteralString(data []byte, cmap *pdfCMap) string {
	buf := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] != '\\' {
			buf = append(buf, data[i])
			continue
		}
		i++
		if i >= len(data) {
			break
		}
		switch data[i] {
		case 'n':
			buf = append(buf, '\n')
		case 'r':
			buf = append(buf, '\r')
		case 't':
			buf = append(buf, '\t')
		case 'b':
			buf = append(buf, '\b')
		case 'f':
			buf = append(buf, '\f')
		case '(', ')', '\\':
			buf = append(buf, data[i])
		case '\n':
		case '\r':
			if i+1 < len(data) && data[i] == '\r' && data[i+1] == '\n' {
				i++
			}
		default:
			if data[i] >= '0' && data[i] <= '7' {
				octal := []byte{data[i]}
				for j := 0; j < 2 && i+1 < len(data) && data[i+1] >= '0' && data[i+1] <= '7'; j++ {
					i++
					octal = append(octal, data[i])
				}
				if v, err := strconv.ParseInt(string(octal), 8, 32); err == nil {
					buf = append(buf, byte(v))
				}
			} else {
				buf = append(buf, data[i])
			}
		}
	}
	if mapped := cmap.decodeBytes(buf); mapped != "" {
		return mapped
	}
	return decodePDFEncodedBytes(buf)
}

func decodePDFHexString(data []byte, cmap *pdfCMap) string {
	cleaned := strings.Map(func(r rune) rune {
		switch {
		case r >= '0' && r <= '9':
			return r
		case r >= 'a' && r <= 'f':
			return r
		case r >= 'A' && r <= 'F':
			return r
		default:
			return -1
		}
	}, string(data))
	if cleaned == "" {
		return ""
	}
	if len(cleaned)%2 == 1 {
		cleaned += "0"
	}
	decoded, err := hex.DecodeString(cleaned)
	if err != nil {
		return ""
	}
	if mapped := cmap.decodeHex(cleaned, decoded); mapped != "" {
		return mapped
	}
	return decodePDFEncodedBytes(decoded)
}

func decodePDFEncodedBytes(data []byte) string {
	if len(data) >= 2 && ((data[0] == 0xFE && data[1] == 0xFF) || (data[0] == 0xFF && data[1] == 0xFE)) {
		u16 := make([]uint16, 0, (len(data)-2)/2)
		be := data[0] == 0xFE
		for i := 2; i+1 < len(data); i += 2 {
			if be {
				u16 = append(u16, uint16(data[i])<<8|uint16(data[i+1]))
			} else {
				u16 = append(u16, uint16(data[i+1])<<8|uint16(data[i]))
			}
		}
		return string(utf16.Decode(u16))
	}
	if looksLikeUTF16BE(data) {
		u16 := make([]uint16, 0, len(data)/2)
		for i := 0; i+1 < len(data); i += 2 {
			u16 = append(u16, uint16(data[i])<<8|uint16(data[i+1]))
		}
		return string(utf16.Decode(u16))
	}
	if looksLikeUTF16LE(data) {
		u16 := make([]uint16, 0, len(data)/2)
		for i := 0; i+1 < len(data); i += 2 {
			u16 = append(u16, uint16(data[i+1])<<8|uint16(data[i]))
		}
		return string(utf16.Decode(u16))
	}
	return string(data)
}

func normalizePDFText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func newPDFCMap() *pdfCMap {
	return &pdfCMap{
		mappingByWidth: make(map[int]map[string]string),
	}
}

func (c *pdfCMap) merge(other map[int]map[string]string) {
	for width, items := range other {
		if len(items) == 0 {
			continue
		}
		if _, ok := c.mappingByWidth[width]; !ok {
			c.mappingByWidth[width] = make(map[string]string)
			c.widths = append(c.widths, width)
		}
		for k, v := range items {
			c.mappingByWidth[width][strings.ToUpper(k)] = v
		}
	}
	for i := 0; i < len(c.widths); i++ {
		for j := i + 1; j < len(c.widths); j++ {
			if c.widths[j] > c.widths[i] {
				c.widths[i], c.widths[j] = c.widths[j], c.widths[i]
			}
		}
	}
}

func (c *pdfCMap) decodeBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return c.decodeHex(strings.ToUpper(hex.EncodeToString(data)), data)
}

func (c *pdfCMap) decodeHex(cleaned string, decoded []byte) string {
	if c == nil || len(c.mappingByWidth) == 0 || cleaned == "" {
		return ""
	}
	for _, width := range c.widths {
		if width <= 0 || len(cleaned)%width != 0 {
			continue
		}
		mapping := c.mappingByWidth[width]
		parts := make([]string, 0, len(cleaned)/width)
		ok := true
		for i := 0; i < len(cleaned); i += width {
			part := cleaned[i : i+width]
			text, exists := mapping[part]
			if !exists {
				ok = false
				break
			}
			parts = append(parts, text)
		}
		if ok && len(parts) > 0 {
			return strings.Join(parts, "")
		}
	}
	return ""
}

func extractPDFCMap(data []byte) map[int]map[string]string {
	result := make(map[int]map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	mode := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasSuffix(line, "beginbfchar"):
			mode = "bfchar"
			continue
		case strings.HasSuffix(line, "endbfchar"):
			mode = ""
			continue
		case strings.HasSuffix(line, "beginbfrange"):
			mode = "bfrange"
			continue
		case strings.HasSuffix(line, "endbfrange"):
			mode = ""
			continue
		}

		switch mode {
		case "bfchar":
			match := pdfBFCharRegexp.FindStringSubmatch(line)
			if len(match) != 3 {
				continue
			}
			addPDFCMapEntry(result, strings.ToUpper(match[1]), decodeUnicodeHex(match[2]))
		case "bfrange":
			match := pdfBFRangeRegexp.FindStringSubmatch(line)
			if len(match) != 6 {
				continue
			}
			startHex := strings.ToUpper(match[1])
			endHex := strings.ToUpper(match[2])
			if match[5] != "" {
				entries := pdfHexTokenRegexp.FindAllStringSubmatch(match[5], -1)
				start, err1 := strconv.ParseInt(startHex, 16, 64)
				end, err2 := strconv.ParseInt(endHex, 16, 64)
				if err1 != nil || err2 != nil {
					continue
				}
				for idx, value := range entries {
					if len(value) < 2 {
						continue
					}
					code := fmt.Sprintf("%0*X", len(startHex), start+int64(idx))
					if int64(idx) > end-start {
						break
					}
					addPDFCMapEntry(result, code, decodeUnicodeHex(value[1]))
				}
				continue
			}

			start, err1 := strconv.ParseInt(startHex, 16, 64)
			end, err2 := strconv.ParseInt(endHex, 16, 64)
			dst, err3 := strconv.ParseInt(match[4], 16, 64)
			if err1 != nil || err2 != nil || err3 != nil {
				continue
			}
			for offset := int64(0); offset <= end-start; offset++ {
				code := fmt.Sprintf("%0*X", len(startHex), start+offset)
				dstHex := fmt.Sprintf("%0*X", len(match[4]), dst+offset)
				addPDFCMapEntry(result, code, decodeUnicodeHex(dstHex))
			}
		}
	}
	return result
}

func addPDFCMapEntry(result map[int]map[string]string, code, text string) {
	if code == "" || text == "" {
		return
	}
	width := len(code)
	if _, ok := result[width]; !ok {
		result[width] = make(map[string]string)
	}
	result[width][strings.ToUpper(code)] = text
}

func decodeUnicodeHex(value string) string {
	cleaned := strings.ToUpper(strings.TrimSpace(value))
	if cleaned == "" {
		return ""
	}
	if len(cleaned)%2 == 1 {
		cleaned = "0" + cleaned
	}
	decoded, err := hex.DecodeString(cleaned)
	if err != nil {
		return ""
	}
	return decodePDFEncodedBytes(decoded)
}

func looksLikeUTF16BE(data []byte) bool {
	if len(data) < 2 || len(data)%2 != 0 {
		return false
	}
	zeroHigh := 0
	for i := 0; i+1 < len(data); i += 2 {
		if data[i] == 0 && data[i+1] != 0 {
			zeroHigh++
		}
	}
	return zeroHigh*2 >= len(data)/2
}

func looksLikeUTF16LE(data []byte) bool {
	if len(data) < 2 || len(data)%2 != 0 {
		return false
	}
	zeroLow := 0
	for i := 0; i+1 < len(data); i += 2 {
		if data[i] != 0 && data[i+1] == 0 {
			zeroLow++
		}
	}
	return zeroLow*2 >= len(data)/2
}
