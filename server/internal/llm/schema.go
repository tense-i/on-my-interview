package llm

import (
	"fmt"
	"strings"
)

func (p *StructuredPost) Validate() error {
	if p == nil {
		return fmt.Errorf("structured post is nil")
	}
	if strings.TrimSpace(p.SchemaVersion) == "" {
		return fmt.Errorf("schema_version is required")
	}
	for idx, question := range p.Questions {
		if strings.TrimSpace(question.Question) == "" {
			return fmt.Errorf("questions[%d].question is required", idx)
		}
		if len(question.Tags) == 0 {
			return fmt.Errorf("questions[%d].tags is required", idx)
		}
		if strings.TrimSpace(question.SourceExcerpt) == "" {
			return fmt.Errorf("questions[%d].source_excerpt is required", idx)
		}
	}
	return nil
}

func FlattenQuestions(rawPostID int64, platform, companyName string, questions []StructuredQuestion) []QuestionRow {
	rows := make([]QuestionRow, 0, len(questions))
	for _, question := range questions {
		rows = append(rows, QuestionRow{
			RawPostID:     rawPostID,
			Platform:      platform,
			CompanyName:   companyName,
			QuestionText:  question.Question,
			QuestionOrder: question.Order,
			Category:      question.Category,
			Tags:          append([]string(nil), question.Tags...),
			SourceExcerpt: question.SourceExcerpt,
		})
	}
	return rows
}
