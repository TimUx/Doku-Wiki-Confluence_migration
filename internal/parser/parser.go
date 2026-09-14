package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

var (
	heading = regexp.MustCompile(`^(={2,6})\s*(.*?)\s*(={2,6})\s*$`)
	link = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
	media = regexp.MustCompile(`\{\{\s*([^}|?]+)(?:\?([^}|]*))?(?:\|([^}]*))?\s*\}\}`)
	include = regexp.MustCompile(`(?i)^\{\{\s*(page|section|namespace|tagtopic|include)>\s*([^}&]+?)(?:&([^}]+))?\s*\}\}$`)
	wrapOpen = regexp.MustCompile(`(?i)^<(WRAP|block|div|wrap|inline|span)\b([^>]*)>(.*)$`)
	wrapSelf = regexp.MustCompile(`(?i)^<(WRAP|block|div|wrap|inline|span)\b([^>]*)/\s*>$`)
	codeOpen = regexp.MustCompile(`(?i)^<(code|file)\b([^>]*)>(.*)$`)
	listItem = regexp.MustCompile(`^(\s*)([*-])\s+(.*)$`)
	tableRow = regexp.MustCompile(`^\s*([|^])(.*)([|^])\s*$`)
	control = regexp.MustCompile(`^~~([A-Z]+)(?:\s+([^~]+))?~~$`)
	numberedHeading = regexp.MustCompile(`^\s*(-{1,2})(?:#([0-9]+)|"([^"]+)")?\s*(.*)$`)
)

type headingState struct { tier1 int; counters [6]int }

func Parse(id, src string) model.Page {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	p := model.Page{ID: id, Source: src, Title: last(id)}
	state := &headingState{}
	parseLines(&p, id, lines, state)
	return p
}

func parseLines(p *model.Page, id string, lines []string, hs *headingState) {
	for i := 0; i < len(lines); i++ {
		raw, trim := lines[i], strings.TrimSpace(lines[i])
		if trim == "" { continue }
		if m := control.FindStringSubmatch(trim); m != nil { p.Nodes = append(p.Nodes, model.Node{Type:"control", Plugin:strings.ToLower(m[1]), Text:strings.TrimSpace(m[2]), Raw:trim}); continue }
		if strings.HasPrefix(trim, "<nowiki>") || strings.HasPrefix(trim, "<NOWIKI>") {
			end := "</nowiki>"; if strings.HasPrefix(trim, "<NOWIKI>") { end = "</NOWIKI>" }
			text := strings.TrimPrefix(trim, trim[:len("<nowiki>")]); var b []string
			if j := strings.Index(strings.ToLower(text), strings.ToLower(end)); j >= 0 { text = text[:j] } else { b = append(b, text); i++; for i < len(lines) && !strings.Contains(strings.ToLower(lines[i]), strings.ToLower(end)) { b = append(b, lines[i]); i++ } }
			if len(b) > 0 { text = strings.Join(b, "\n") }; p.Nodes = append(p.Nodes, model.Node{Type:"nowiki", Text:text}); continue
		}
		if m := heading.FindStringSubmatch(trim); m != nil && len(m[1]) == len(m[3]) {
			level := 7-len(m[1]); text := strings.TrimSpace(m[2]); text, numbered := numberHeading(text, level, hs)
			if numbered { p.Plugins = appendUnique(p.Plugins, "numberedheadings") }
			p.Nodes = append(p.Nodes, model.Node{Type:"heading", Text:text, Level:strconv.Itoa(level)}); if p.Title == last(id) { p.Title = stripNumberPrefix(text) }; continue
		}
		if m := codeOpen.FindStringSubmatch(trim); m != nil {
			tag, attrs, rest := strings.ToLower(m[1]), strings.TrimSpace(m[2]), m[3]; end := "</"+tag+">"; meta := parseCodeAttrs(attrs); var b []string
			if j := strings.Index(strings.ToLower(rest), end); j >= 0 { b = append(b, rest[:j]) } else { i++; if strings.TrimSpace(rest) != "" { b = append(b, rest) }; for i < len(lines) && strings.TrimSpace(strings.ToLower(lines[i])) != end { b = append(b, lines[i]); i++ } }
			p.Nodes = append(p.Nodes, model.Node{Type:"code", Text:strings.TrimRight(strings.Join(b,"\n"),"\n"), Meta:meta}); continue
		}
		if m := wrapSelf.FindStringSubmatch(trim); m != nil { p.Plugins=appendUnique(p.Plugins,"wrap"); p.Nodes=append(p.Nodes, wrapNode(m[1],m[2],nil,trim)); continue }
		if m := wrapOpen.FindStringSubmatch(trim); m != nil {
			name := m[1]; lower := strings.ToLower(name); if lower == "wrap" { p.Plugins=appendUnique(p.Plugins,"wrap") }
			depth := 1; var body []string; rest := m[3]; closeTag := "</"+name+">"
			if j := strings.Index(strings.ToLower(rest), strings.ToLower(closeTag)); j >= 0 { body = append(body, rest[:j]) } else { i++; if strings.TrimSpace(rest) != "" { body=append(body,rest) }; for i<len(lines) { line:=lines[i]; if hasWrapOpen(line,name) { depth++ }; if hasWrapClose(line,name) { depth--; if depth==0 { if j:=strings.Index(strings.ToLower(line),strings.ToLower(closeTag));j>0{body=append(body,line[:j])}; break } }; body=append(body,line); i++ } }
			sub := &model.Page{ID:id,Title:p.Title}; parseLines(sub,id,body,hs); mergeRefs(p,sub); p.Plugins=appendUnique(p.Plugins,"wrap"); n:=wrapNode(name,m[2],sub.Nodes,trim); p.Nodes=append(p.Nodes,n); continue
		}
		if m := include.FindStringSubmatch(trim); m != nil { kind:=strings.ToLower(m[1]); target:=resolve(id,unescape(m[2])); meta:=map[string]string{"mode":kind,"flags":strings.TrimSpace(m[3])}; section:=""; if strings.HasPrefix(target,"#") { section=strings.TrimPrefix(target,"#") }; if section!="" { meta["section"]=section }; p.Includes=append(p.Includes,model.Reference{Target:target,Display:section,Kind:kind}); p.Plugins=appendUnique(p.Plugins,"include"); p.Nodes=append(p.Nodes,model.Node{Type:"include",Target:target,Raw:trim,Meta:meta}); continue }
		if m := tableRow.FindStringSubmatch(trim); m != nil { if row:=parseTableRow(m[1],m[2]); len(row.Children)>0 { p.Nodes=append(p.Nodes,row); continue } }
		if m := listItem.FindStringSubmatch(raw); m != nil { typ:="bullet_item"; if m[2]=="-" { typ="ordered_item" }; indent:=len(strings.ReplaceAll(m[1],"\t","    ")); p.Nodes=append(p.Nodes,model.Node{Type:typ,Text:m[3],Meta:map[string]string{"indent":strconv.Itoa(indent)}}); continue }
		if strings.HasPrefix(trim, ">") { level:=0; for level<len(trim)&&trim[level]=='>' { level++ }; text:=strings.TrimSpace(strings.TrimLeft(trim,">")); p.Nodes=append(p.Nodes,model.Node{Type:"quote",Text:text,Meta:map[string]string{"level":strconv.Itoa(level)}}); continue }
		if trim == "----" || strings.Trim(trim,"-")=="" && len(trim)>=4 { p.Nodes=append(p.Nodes,model.Node{Type:"hr",Raw:trim}); continue }
		if strings.Contains(raw,"[[") || strings.Contains(raw,"{{") { collectRefs(p,id,raw) }
		p.Nodes=append(p.Nodes,model.Node{Type:"paragraph",Text:trim})
	}
}

func numberHeading(text string, level int, hs *headingState) (string,bool) {
	m:=numberedHeading.FindStringSubmatch(text); if m==nil || m[1]=="" { return text,false }
	if hs.tier1==0 { hs.tier1=2 }; idx:=level-hs.tier1; if idx<0 { idx=0 }; if idx>5 { idx=5 }
	for i:=idx+1;i<6;i++ { hs.counters[i]=0 }; if m[2]!="" { hs.counters[idx]=atoi(m[2]) } else { hs.counters[idx]++ }
	parts:=[]string{}; for i:=0;i<=idx;i++ { if hs.counters[i]>0 { parts=append(parts,strconv.Itoa(hs.counters[i])) } }; title:=strings.TrimSpace(m[4]); if m[3]!="" { title=strings.TrimSpace(m[3])+" "+title }; if title=="" { return strings.Join(parts,"."),true }; return strings.Join(parts,".")+" "+title,true
}
func stripNumberPrefix(s string) string { return s }
func atoi(s string) int { n,_:=strconv.Atoi(s); return n }
func parseCodeAttrs(s string) map[string]string { m:=map[string]string{}; f:=strings.Fields(s); if len(f)>0 && !strings.Contains(f[0],"="){m["language"]=f[0]}; if len(f)>1 && !strings.Contains(f[1],"="){m["filename"]=f[1]}; for _,x:=range f { if i:=strings.Index(x,"=");i>0 {m[x[:i]]=strings.Trim(x[i+1:],"\"")} }; return m }
func wrapNode(name, attrs string, children []model.Node, raw string) model.Node { meta:=parseWrapAttrs(attrs); meta["tag"]=strings.ToLower(name); return model.Node{Type:"wrap",Plugin:"wrap",Children:children,Meta:meta,Raw:raw} }
func parseWrapAttrs(s string) map[string]string { m:=map[string]string{}; s=strings.TrimSpace(s); if s=="" {return m}; classes:=[]string{}; for _,f:=range strings.Fields(s) { if strings.HasPrefix(f,"#"){m["id"]=strings.TrimPrefix(f,"#")} else if strings.HasPrefix(f,":"){m["language"]=strings.TrimPrefix(f,":")} else if strings.Contains(f,"="){p:=strings.SplitN(f,"=",2);m[p[0]]=strings.Trim(p[1],"\"")} else {classes=append(classes,f)} }; if len(classes)>0 {m["classes"]=strings.Join(classes," ")}; return m }
func hasWrapOpen(s, name string) bool { re:=regexp.MustCompile(`(?i)<`+regexp.QuoteMeta(name)+`\b[^>]*>`); return re.MatchString(s) }
func hasWrapClose(s, name string) bool { return strings.Contains(strings.ToLower(s),strings.ToLower("</"+name+">")) }
func parseTableRow(sep, body string) model.Node { if strings.HasSuffix(body,sep){body=strings.TrimSuffix(body,sep)}; parts:=strings.Split(body,sep); row:=model.Node{Type:"table_row"}; for _,part:=range parts { cell:=strings.TrimSpace(part); if cell=="" {continue}; meta:=map[string]string{}; if cell==":::" { if len(row.Children)>0 {row.Children[len(row.Children)-1].Meta["colspan"]=strconv.Itoa(atoi(row.Children[len(row.Children)-1].Meta["colspan"])+1); continue } }; typ:="table_cell"; if sep=="^" {typ="table_header"}; row.Children=append(row.Children,model.Node{Type:typ,Text:cell,Meta:meta}) }; return row }
func collectRefs(p *model.Page,id,raw string) { for _,m:=range link.FindAllStringSubmatch(raw,-1) { target:=unescape(m[1]); kind:="internal"; if isExternal(target){kind="external"} else if strings.HasPrefix(strings.TrimSpace(target),"#"){kind="same-page"} else {target=resolve(id,target)}; p.Links=append(p.Links,model.Reference{Target:target,Display:unescape(m[2]),Kind:kind}); if m[2]!="" { for _,mm:=range media.FindAllStringSubmatch(m[2],-1){t:=normalizeMediaTarget(mm[1]);p.Media=appendUniqueRef(p.Media,model.Reference{Target:resolve(id,t),Display:unescape(mm[3]),Kind:"media"})} } }; for _,m:=range media.FindAllStringSubmatch(raw,-1){target:=normalizeMediaTarget(m[1]);p.Media=appendUniqueRef(p.Media,model.Reference{Target:resolve(id,target),Display:unescape(m[3]),Kind:"media"})} }
func mergeRefs(dst,src *model.Page) { for _,r:=range src.Links {dst.Links=append(dst.Links,r)}; for _,r:=range src.Media {dst.Media=appendUniqueRef(dst.Media,r)}; for _,r:=range src.Includes {dst.Includes=append(dst.Includes,r)}; for _,x:=range src.Plugins {dst.Plugins=appendUnique(dst.Plugins,x)}; dst.Warnings=append(dst.Warnings,src.Warnings...) }
func appendUniqueRef(s []model.Reference,v model.Reference)[]model.Reference{for _,x:=range s{if x.Target==v.Target&&x.Kind==v.Kind{return s}};return append(s,v)}
func normalizeMediaTarget(target string)string{target=unescape(strings.TrimSpace(target));if i:=strings.Index(target,"fetch.php/");i>=0{target=target[i+len("fetch.php/"):]} ;return strings.TrimPrefix(target,"/")}
func unescape(s string)string{for _,pair:=range []struct{escaped,plain string}{{`\:`,":"},{`\_`,`_`},{`\.`,`.`},{`\-`,`-`},{`\+`,`+`},{`\#`,`#`},{`\&`,`&`},{`\?`,`?`},{`\|`,`|`},{`\*`,`*`}}{s=strings.ReplaceAll(s,pair.escaped,pair.plain)};s=strings.ReplaceAll(s,"https:*","https://");s=strings.ReplaceAll(s,"http:*","http://");return s}
func isExternal(s string)bool{s=unescape(strings.TrimSpace(s));return strings.Contains(s,"://")||strings.HasPrefix(strings.ToLower(s),"mailto:")}
func resolve(current,target string)string{target=strings.TrimSpace(target);if strings.HasPrefix(target,":"){return strings.TrimPrefix(target,":")};if strings.HasPrefix(target,"#"){return target};if strings.Contains(target,":"){return target};if i:=strings.LastIndex(current,":");i>=0{return current[:i+1]+target};return target}
func last(id string)string{if i:=strings.LastIndex(id,":");i>=0{return id[i+1:]};return id}
func appendUnique(s []string,v string)[]string{for _,x:=range s{if x==v{return s}};return append(s,v)}
