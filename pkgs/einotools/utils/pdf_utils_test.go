package utils

import (
	"strings"
	"testing"
)

func TestReadPDFLinesRange(t *testing.T) {
	pdfContent := `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 55 >>
stream
BT
/F1 12 Tf
72 720 Td
(Hello PDF) Tj
0 -20 Td
(Second line) Tj
ET
endstream
endobj
trailer
<< /Root 1 0 R >>
%%EOF`

	lines, err := ReadPDFLinesRange(strings.NewReader(pdfContent), 1, -1)
	if err != nil {
		t.Fatalf("ReadPDFLinesRange failed: %v", err)
	}
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "Hello PDF" {
		t.Fatalf("unexpected first line: %q", lines[0])
	}
	if lines[1] != "Second line" {
		t.Fatalf("unexpected second line: %q", lines[1])
	}
}

func TestReadPDFLinesRange_WithToUnicodeCMap(t *testing.T) {
	pdfContent := `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 80 >>
stream
BT
/F1 12 Tf
72 720 Td
<0001000200030004> Tj
ET
endstream
endobj
5 0 obj
<< /Length 220 >>
stream
/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
1 beginbfchar
<0001> <0032>
<0002> <0030>
<0003> <0032>
<0004> <0035>
endbfchar
endcmap
CMapName currentdict /CMap defineresource pop
end
end
endstream
endobj
trailer
<< /Root 1 0 R >>
%%EOF`

	lines, err := ReadPDFLinesRange(strings.NewReader(pdfContent), 1, -1)
	if err != nil {
		t.Fatalf("ReadPDFLinesRange failed: %v", err)
	}
	if len(lines) == 0 {
		t.Fatalf("expected at least 1 line")
	}
	if lines[0] != "2025" {
		t.Fatalf("unexpected first line: %q", lines[0])
	}
}

func TestReadPDFLinesRange_WithTJArray(t *testing.T) {
	pdfContent := `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 96 >>
stream
BT
/F1 12 Tf
72 720 Td
[<0001><0002><0003><0004><0005><0006><0007><0008><0009><000A>] TJ
ET
endstream
endobj
5 0 obj
<< /Length 340 >>
stream
/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
10 beginbfchar
<0001> <0032>
<0002> <0030>
<0003> <0032>
<0004> <0035>
<0005> <0031>
<0006> <0031>
<0007> <0030>
<0008> <0031>
<0009> <0030>
<000A> <0032>
endbfchar
endcmap
CMapName currentdict /CMap defineresource pop
end
end
endstream
endobj
trailer
<< /Root 1 0 R >>
%%EOF`

	lines, err := ReadPDFLinesRange(strings.NewReader(pdfContent), 1, -1)
	if err != nil {
		t.Fatalf("ReadPDFLinesRange failed: %v", err)
	}
	if len(lines) == 0 {
		t.Fatalf("expected at least 1 line")
	}
	if lines[0] != "2025110102" {
		t.Fatalf("unexpected first line: %q", lines[0])
	}
}
