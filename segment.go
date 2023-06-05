package sqls

// Segment is the builder for a part of or even the full query, it allows you
// to write and combine segments with freedom.
type Segment struct {
	Raw      string         // Raw string support bindvars and preprocessing functions.
	Args     []any          // Args to be referenced by the Raw
	Columns  []*TableColumn // Columns to be referenced by the Raw
	Tables   []Table        // Table names / alias to be referenced by the Raw
	Segments []*Segment     // Segments to be referenced by the Raw
	Builders []Builder      // Builders to be referenced by the Raw

	Prefix string // Prefix is added before the rendered segment only if which is not empty.
	Suffix string // Suffix is added after the rendered segment only if which is not empty.
}

// AppendTables appends tables to the segment.
func (s *Segment) AppendTables(tables ...Table) {
	s.Tables = append(s.Tables, tables...)
}

// AppendColumns appends columns to the segment.
func (s *Segment) AppendColumns(columns ...*TableColumn) {
	s.Columns = append(s.Columns, columns...)
}

// AppendSegments appends segments to the s.Segments.
func (s *Segment) AppendSegments(segments ...*Segment) {
	s.Segments = append(s.Segments, segments...)
}

// AppendArgs appends args to the s.Args.
func (s *Segment) AppendArgs(args ...any) {
	s.Args = append(s.Args, args...)
}

// WithTables replace s.Tables with the tables
func (s *Segment) WithTables(tables ...Table) {
	s.Tables = tables
}

// WithColumns replace s.Columns with the columns
func (s *Segment) WithColumns(columns ...*TableColumn) {
	s.Columns = columns
}

// WithSegments replace s.Segments with the segments
func (s *Segment) WithSegments(segments ...*Segment) {
	s.Segments = segments
}

// WithArgs replace s.Args with the args
func (s *Segment) WithArgs(args ...any) {
	s.Args = args
}
