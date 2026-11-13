package docsearcher

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html/charset"
)

type Book struct {
	Metadata TitleInfo
	Text     string
}

type TitleInfo struct {
	Genres     []string `xml:"genre"`
	Authors    []Author `xml:"author"`
	BookTitle  string   `xml:"book-title"`
	Annotation struct {
		Text string `xml:",innerxml"`
	} `xml:"annotation"`
	Lang string `xml:"lang"`
}

type Author struct {
	FirstName  string `xml:"first-name"`
	MiddleName string `xml:"middle-name"`
	LastName   string `xml:"last-name"`
}

func ReadFB2(rc io.ReadCloser) (*Book, error) {
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	decoder.CharsetReader = charset.NewReaderLabel

	var book Book
	var foundMeta bool
	var inBody bool
	var textBuilder strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "title-info":
				if !foundMeta {
					if err := decoder.DecodeElement(&book.Metadata, &t); err != nil {
						return nil, err
					}
					foundMeta = true
				}
			case "body":
				isNotes := false
				for _, attr := range t.Attr {
					if attr.Name.Local == "name" && attr.Value == "notes" {
						isNotes = true
						break
					}
				}
				if !isNotes {
					inBody = true
				}
			}

		case xml.EndElement:
			switch t.Name.Local {
			case "body":
				inBody = false
			case "p", "v", "subtitle", "empty-line":
				if inBody {
					textBuilder.WriteString("\n")
				}
			}

		case xml.CharData:
			if inBody {
				cleanText := strings.ReplaceAll(string(t), "\n", " ")
				cleanText = strings.ReplaceAll(cleanText, "\r", "")
				textBuilder.WriteString(cleanText)
			}
		}
	}

	if !foundMeta {
		return nil, fmt.Errorf("tag title-info not found")
	}

	book.Text = strings.TrimSpace(textBuilder.String())
	return &book, nil
}
