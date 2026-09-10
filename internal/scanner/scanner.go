package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/parser"
)

type Result struct { Pages []model.Page; Media []model.Media; Warnings []model.Warning }

func Scan(ctx context.Context,pagesRoot,mediaRoot string)(Result,error){
	var out Result
	err:=filepath.WalkDir(pagesRoot,func(path string,d fs.DirEntry,err error)error{if err!=nil{out.Warnings=append(out.Warnings,model.Warning{Code:"read",Message:err.Error()});return nil};select{case<-ctx.Done():return ctx.Err();default:};if d.IsDir()||!strings.HasSuffix(strings.ToLower(d.Name()),".txt"){return nil};rel,_:=filepath.Rel(pagesRoot,path);id:=strings.TrimSuffix(filepath.ToSlash(rel),filepath.Ext(rel));id=strings.ReplaceAll(id,"/",":");b,e:=os.ReadFile(path);if e!=nil{out.Warnings=append(out.Warnings,model.Warning{Code:"read",Message:fmt.Sprintf("cannot read %s",id)});return nil};info,_:=d.Info();p:=parser.Parse(id,string(b));p.SourceFile=path;p.Namespace=namespace(id);p.Size=info.Size();p.Modified=info.ModTime();out.Pages=append(out.Pages,p);return nil});if err!=nil{return out,err}
	err=filepath.WalkDir(mediaRoot,func(path string,d fs.DirEntry,err error)error{if err!=nil{return nil};if d.IsDir(){return nil};rel,_:=filepath.Rel(mediaRoot,path);info,_:=d.Info();out.Media=append(out.Media,model.Media{ID:strings.ReplaceAll(filepath.ToSlash(rel),"/",":"),Path:path,Size:info.Size(),Modified:info.ModTime()});return nil});return out,err
}
func namespace(id string)string{if i:=strings.LastIndex(id,":");i>=0{return id[:i]};return ""}
