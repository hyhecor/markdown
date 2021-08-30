package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/suapapa/go_hangul/encoding/cp949"
)

func main() {

	//읽기
	encoded_cp949, err := ReadFull(os.Stdin)
	Check(err)

	//csv파일은 기본 인코딩 cp494
	//한글 깨져서 utf-8 변경해야 함
	r := bytes.NewReader(encoded_cp949)
	cp949_decoder, err := cp949.NewReader(r)
	Check(err)

	plane_text := make([]byte, len(encoded_cp949)<<1)
	n, err := cp949_decoder.Read(plane_text)
	Check(err)

	mdtbl, err := CvtCsvToMarkdowntable(plane_text[:n])
	Check(err)

	w := bufio.NewWriter(os.Stdout)
	_, err = w.Write(mdtbl)
	Check(err)
	w.Flush()
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

const RecoveringDefaultFormatString string = "Recovering from panic:%v"

func Recovering(fn func(), handlers ...func(r interface{})) {

	defer func() {
		if r := recover(); r != nil {
			for _, handling := range handlers {
				handling(r)
			}
		}
	}()

	fn()
}

//ReadFull EOF가 올때까지 계속 읽는다
func ReadFull(r io.Reader) (buff []byte, err error) {

	buff = make([]byte, 0, 1<<10)
	b := make([]byte, 1)
	_, err = io.ReadFull(r, b)
	for err == nil {
		buff = append(buff, b...)
		_, err = io.ReadFull(r, b)
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func CvtCsvToMarkdowntable(csvData []byte) (buff []byte, err error) {
	const (
		text_newline              = "\n"
		html_newline              = "<br>"
		text_space                = " "
		html_space                = "&nbsp;"
		markdown_column_seperater = "|"
		markdown_head_seperater   = "|---"
	)

	r := bytes.NewReader(csvData)
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return
	}

	rowFold := func(row []string, delim string) string {
		var out string
		for _, s := range row {
			out += delim + strings.ReplaceAll(strings.ReplaceAll(s, text_newline, html_newline), text_space, html_space)
		}
		return out
	}

	seperaterFold := func(row []string, delim string) string {
		return strings.Repeat(delim, len(row))
	}

	buff = make([]byte, 0, 1<<10)
	for i, rows := range records {

		s := fmt.Sprintln(rowFold(rows, markdown_column_seperater))
		if i == 0 {
			s += fmt.Sprintln(seperaterFold(rows, markdown_head_seperater))
		}
		buff = append(buff, []byte(s)...)
	}

	return
}
