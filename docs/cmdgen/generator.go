// cmdgen is a tool for generating pages from a cobra command for a Hugo site.
package cmdgen

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/hugo/helpers"
	"github.com/spf13/hugo/hugolib"
	"github.com/spf13/hugo/parser"
)

type CmdDocsGeneratorOptions struct {
	BasePath           string
	MetadataAdditional map[string]interface{}
	Menus              hugolib.PageMenus
}

type CmdDocsGenerator struct {
	dir       string
	opts      CmdDocsGeneratorOptions
	indexLink string
}

func GenerateCmdDocs(cmd *cobra.Command, dir string, opts CmdDocsGeneratorOptions) error {
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	s, err := hugolib.NewEnglishSite()
	if err != nil {
		return err
	}
	g := NewCmdDocsGenerator(dir, opts)
	if err := g.GenMarkdownTree(s, cmd); err != nil {
		return err
	}
	return nil
}

func NewCmdDocsGenerator(dir string, opts CmdDocsGeneratorOptions) *CmdDocsGenerator {
	return &CmdDocsGenerator{dir, opts, ""}
}

func (g *CmdDocsGenerator) GenMarkdownTree(s *hugolib.Site, cmd *cobra.Command) error {
	return g.genMarkdownTree(s, cmd, 0)
}

func (g *CmdDocsGenerator) genMarkdownTree(s *hugolib.Site, cmd *cobra.Command, depth int) error {
	pagename := cmd.CommandPath()
	basename := strings.Replace(pagename, " ", "_", -1) + ".md"
	if depth == 0 {
		g.indexLink = basename
		basename = "index.md"
	}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := g.genMarkdownTree(s, c, depth+1); err != nil {
			return err
		}
	}

	filename := s.PathSpec.AbsPathify(filepath.Join(g.dir, basename))
	page, err := s.NewPage(pagename)
	if err != nil {
		return err
	}
	if err := page.SetSourceMetaData(g.createMetadata(cmd, depth), parser.FormatToLeadRune(s.Cfg.GetString("metaDataFormat"))); err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	if err := doc.GenMarkdownCustom(cmd, buf, g.linkHandler); err != nil {
		return err
	}

	page.SetSourceContent(buf.Bytes())

	if err := page.SafeSaveSourceAs(filename); err != nil {
		return err
	}
	log.Println(filename, "created")
	return nil
}

func (g *CmdDocsGenerator) createMetadata(cmd *cobra.Command, depth int) map[string]interface{} {
	metadata := make(map[string]interface{})
	metadata["title"] = helpers.MakeTitle(cmd.CommandPath())
	metadata["date"] = time.Now().Format(time.RFC3339)
	for k, v := range g.opts.MetadataAdditional {
		metadata[k] = v
	}
	if depth == 0 {
		for k, v := range pageMenusUnmarshalMap(g.opts.Menus) {
			metadata[k] = v
		}
	}
	return metadata
}

func (g *CmdDocsGenerator) linkHandler(link string) string {
	if link == g.indexLink {
		link = ""
	}
	return filepath.Join(g.opts.BasePath, link[:len(link)-len(filepath.Ext(link))]) + "/"
}

func pageMenusUnmarshalMap(pm hugolib.PageMenus) (m map[string]interface{}) {
	m = make(map[string]interface{})
	for k, me := range pm {
		mk := make(map[string]interface{})
		if me.URL != "" {
			mk["url"] = me.URL
		}
		mk["weight"] = me.Weight
		if me.Name != "" {
			mk["name"] = me.Name
		}
		if me.Identifier != "" {
			mk["identifier"] = me.Identifier
		}
		if me.Parent != "" {
			mk["parent"] = me.Parent
		}
		m[k] = mk
	}
	return
}
