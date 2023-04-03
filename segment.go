package sqls

var _ Builder = (*Segment)(nil)

// Segment is the builder for a part of or even the full query, it allows you
// to write with placeholders, and combine any segments with freedom.
//
// With the help of Segment, we pay attention only to the reference relationships
// inside the segment, for example, use "$1" to refer the first element of s.Args.
//
// Bindvars supported:
//   - $1, $2 ...
//   - ?, ?, ...
//
// Built-in preprocessors:
//   - #c, #col, #column: refer to the column by index, e.g. #c1, #c(1)
//   - #t, #table: refer to the table by index, e.g. #t1, #t(1)
//   - #s, #seg, #segment: refer to the segment by index, e.g. #s1, #s(1)
//   - #join: join the template (with references) by the separator, e.g. #join('#column', ', '), #join('#c=#$', ', ')
//   - #$: refer to the argument by index, used in #join().
//   - #?: refer to the argument by index, used in #join().
//
// Note, references in the #join template are funcs, with no arguments.
type Segment struct {
	Header   string         // Header is added before the segment only if the body is not empty.
	Footer   string         // Footer is added after the segment only if the body is not empty.
	Raw      string         // Raw string support placeholders
	Segments []*Segment     // Segments referenced by the Raw
	Columns  []*TableColumn // Columns referenced by the Raw
	Args     []any          // Args referenced by the Raw
}

// AppendColumns appends columns to the segment.
func (s *Segment) AppendColumns(columns ...*TableColumn) *Segment {
	s.Columns = append(s.Columns, columns...)
	return s
}

// AppendSegments appends segments to the s.Segments.
func (s *Segment) AppendSegments(segments ...*Segment) *Segment {
	s.Segments = append(s.Segments, segments...)
	return s
}

// AppendArgs appends args to the s.Args.
func (s *Segment) AppendArgs(args ...any) *Segment {
	s.Args = append(s.Args, args...)
	return s
}
