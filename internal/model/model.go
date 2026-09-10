package model

import "time"

type Node struct { Type, Text, Target, Display, Level, Plugin, Raw string; Children []Node; Meta map[string]string }
type Reference struct { Target, Display, Kind string }
type Warning struct { Code, Message, Raw string; Line int }
type Page struct {
	ID, Namespace, Title, SourceFile, SourceURL, Source string
	Size int64; Modified time.Time; Nodes []Node; Links, Media, Includes []Reference; Plugins []string; Warnings []Warning
}
type Media struct { ID, Path, MIME string; Size int64; Modified time.Time }
