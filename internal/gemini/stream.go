package gemini

import (
	"errors"
	"fmt"
	"strings"
)

type StreamParser struct {
	prevText string
	buf      string
}

func NewStreamParser() *StreamParser {
	return &StreamParser{}
}

func (p *StreamParser) Reset() {
	p.buf = ""
}

func (p *StreamParser) Feed(chunk string) ([]string, error) {
	p.buf += chunk

	if strings.Contains(p.buf, "BardErrorInfo") {
		if code, ok := IsBardError(p.buf); ok {
			return nil, fmt.Errorf("Gemini upstream rejected request: BardErrorInfo [%s]", code)
		}
	}

	var deltas []string
	for strings.Contains(p.buf, "\n") {
		idx := strings.Index(p.buf, "\n")
		line := p.buf[:idx]
		p.buf = p.buf[idx+1:]

		texts := ExtractTextsFromLine(line)
		for _, t := range texts {
			if t == p.prevText || strings.HasPrefix(p.prevText, t) {
				continue
			}

			if !strings.HasPrefix(t, p.prevText) {
				return nil, errors.New("Gemini stream content changed during retry")
			}

			delta := CleanText(t[len(p.prevText):], false)
			p.prevText = t
			if delta != "" {
				deltas = append(deltas, delta)
			}
		}
	}

	return deltas, nil
}
