package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/schema"
)

type Row struct {
	Timestamp int64
	Name      []byte
	Status    int32
}

type Rows []*Row

func (rs Rows) Timestamps() []int64 {
	v := make([]int64, 0, len(rs))
	for _, r := range rs {
		v = append(v, r.Timestamp)
	}
	return v
}

func (rs Rows) Names() []parquet.ByteArray {
	v := make([]parquet.ByteArray, 0, len(rs))
	for _, r := range rs {
		v = append(v, r.Name)
	}
	return v
}
func (rs Rows) Statuses() []int32 {
	v := make([]int32, 0, len(rs))
	for _, r := range rs {
		v = append(v, r.Status)
	}
	return v
}

var t *time.Time

func getTimestamp(sec int) int64 {
	if t == nil {
		tt := time.Now()
		t = &tt
	}
	ts := t.Add(time.Duration(sec) * time.Second).UnixMilli()
	return ts
}

var rows = Rows{}

func init() {
	for i := 0; i < 10000000; i++ {
		rows = append(rows, &Row{
			getTimestamp(i),
			[]byte(fmt.Sprintf("test%d", i)),
			int32(100 + i),
		})
	}
}

func main() {
	fields := make(schema.FieldList, 0)
	n, _ := schema.NewPrimitiveNodeLogical("timestamp", parquet.Repetitions.Required, schema.NewTimestampLogicalType(false, schema.TimeUnitMillis), parquet.Types.Int64, 0, -1)
	fields = append(fields, n)
	fields = append(fields, schema.NewByteArrayNode("name", parquet.Repetitions.Required, -1))
	fields = append(fields, schema.NewInt32Node("status", parquet.Repetitions.Required, -1))

	schema, err := schema.NewGroupNode("schema", parquet.Repetitions.Required, fields, -1)
	if err != nil {
		panic(err)
	}

	f, err := os.Create("test.parquet")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := file.NewParquetWriter(f, schema, file.WithWriterProps(parquet.NewWriterProperties(
		parquet.WithPageIndexEnabled(true),
		parquet.WithCompression(compress.Codecs.Snappy),
	)))
	defer w.Close()

	colIndex := 0
	rgr := w.AppendRowGroup()
	nextColumn := func() file.ColumnChunkWriter {
		defer func() {
			colIndex++
		}()

		cw, _ := rgr.NextColumn()
		return cw
	}

	timestampWriter := nextColumn().(*file.Int64ColumnChunkWriter)

	var offset int64
	offset, err = timestampWriter.WriteBatch(rows.Timestamps(), nil, nil)
	if err != nil {
		panic(err)
	}
	log.Println("timestamp write", offset, "rows")

	nameWriter := nextColumn().(*file.ByteArrayColumnChunkWriter)
	offset, err = nameWriter.WriteBatch(rows.Names(), nil, nil)
	if err != nil {
		panic(err)
	}
	log.Println("name write", offset, "rows")

	statusWriter := nextColumn().(*file.Int32ColumnChunkWriter)
	offset, err = statusWriter.WriteBatch(rows.Statuses(), nil, nil)
	if err != nil {
		panic(err)
	}
	log.Println("status write", offset, "rows")
}
