package page

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

var mapLock sync.Mutex

// Render is the main type for this package. 
// Create a variable of this type and specify its fields, then you have 
// access to Show and String functions.
type Render struct {
	TemplateDir string                        // Path to templates.
	Functions   template.FuncMap              // A map of functions we want to pass to our templates.
	UseCache    bool                          // If true, use the template cache, stored in TemplateMap.
	TemplateMap map[string]*template.Template // Our template cache.
	Partials    []string                      // A list of partials.
	Debug       bool                          // Prints debugging info when true.
}

// New returns a Render type populated with sensible defaults.
func New() *Render {
	return &Render{
		Functions:   template.FuncMap{},
		UseCache:    true,
		TemplateMap: make(map[string]*template.Template),
		Partials:    []string{},
		Debug:       false,
	}
}

// Show generates an HTML page from template file(s).
// @ t:
// -	template name: "home.page.tmpl"
// @ td:
// -	template data: 
//			data := make(map[string]any)
//			data["payload"] = "This is MY passed data."
func (ren *Render) Show(w http.ResponseWriter, t string, td any) error {
	// Call buildTemplate to get the template, either from the cache or by building it from disk.
	tmpl, err := ren.buildTemplate(t)
	if err != nil {
		log.Println("error building", err)
		return err
	}

	// Execute the template.
	if err := tmpl.ExecuteTemplate(w, t, td); err != nil {
		log.Println("error executing", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// String renders a template and returns it as a string.
func (ren *Render) String(t string, td any) (string, error) {
	// Call buildTemplate to get the template, either from the cache or by building it
	// from disk.
	tmpl, err := ren.buildTemplate(t)
	if err != nil {
		return "", err
	}

	// Execute the template, storing the result in a bytes.Buffer variable.
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, td); err != nil {
		return "", err
	}

	// Return a string from the bytes.Buffer.
	result := tpl.String()
	return result, nil
}

// GetTemplate attempts to get a template from cache -
//	builds it if it does not find it - and returns it.
func (ren *Render) GetTemplate(t string) (*template.Template, error) {
	// Call buildTemplate to get the template, either from the cache or by building it
	// from disk.
	tmpl, err := ren.buildTemplate(t)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// buildTemplate a utility function that creates a template, 
//	either from cache, or from disk. 
//	The template is ready to accept functions & data, and then get rendered.
// @ t:
// -	template name: "home.page.tmpl"
// @ return:
// -	an actually executable template set
func (ren *Render) buildTemplate(t string) (*template.Template, error) {
	// tmpl is the variable that will hold our template set
	var tmpl *template.Template

	// If we are using the cache, get try to get the pre-compiled template from our
	// map templateMap, stored in the receiver.
	if ren.UseCache {
		if templateFromMap, ok := ren.TemplateMap[t]; ok {
			if ren.Debug {
				log.Println("Reading template", t, "from cache")
			}
			tmpl = templateFromMap
		}
	}

	// At this point, tmpl will be nil if we do not have a value in the map (our template
	// cache). In this case, we build the template from disk.
	if tmpl == nil {
		log.Println("t", t)
		newTemplate, err := ren.buildTemplateFromDisk(t)
		if err != nil {
			log.Println("Error building from disk")
			return nil, err
		}
		tmpl = newTemplate
	}

	return tmpl, nil
}

// buildTemplateFromDisk builds a new template set from disk.
// @ return:
// -	an actually executable template set
func (ren *Render) buildTemplateFromDisk(t string) (*template.Template, error) {
	fmt.Println("139 - page-buildTemplateFromDisk.t: ", t)
	// 139 - page-buildTemplateFromDisk.t:  home.page.tmpl
	// 't' becomes the name of the (future) template set.
	// the key in map[string]*.template.Template

	// templateSlice will hold all templates (names / file names) necessary to 
	// build a finished template set.
	var templateSlice []string

	// Read in the partials, if any.
	// Read any partial associated with this (future) template set.
	// 'Future' because this is still a bunch of text.
	templateSlice = append(templateSlice, ren.Partials...)

	// Append the template name we want to render to the slice. 
	// Use path.Join to make it os agnostic.
	templateSlice = append(templateSlice, path.Join(ren.TemplateDir, t))

	// Create a new template set by parsing all files in the slice.
	tmpl, err := template.New(t).Funcs(ren.Functions).ParseFiles(templateSlice...)
	if err != nil {
		return nil, err
	}

	// Add the template set to the template map stored in our receiver.
	// Note that this(?) is ignored in development, but does not hurt anything.
	// Well, I trust it's not ignored. Otherwise there would be no template set
	// in the map.
	// So here is the template set 'tmpl' added: map["home.page.tmpl"] = tmpl
	mapLock.Lock()
	ren.TemplateMap[t] = tmpl
	mapLock.Unlock()

	// show the contents of map["home.page.tmpl"]
	tpl := ren.TemplateMap["about.page.tmpl"]
	fmt.Println("174 - page-tpl.DefinedTemplates(): ", tpl.DefinedTemplates())
	// 139 - page-buildTemplateFromDisk.t:  home.page.tmpl
	// 174 - page-tpl.DefinedTemplates():  ; 
	//		defined templates are: "css", "title", "css.partial.tmpl", "footer.partial.tmpl", 
	//													 "home.page.tmpl", "content", "footer", "base", 
	//													 "base.layout.tmpl", "title.partial.tmpl"
	// So, all this is available when we ender "home.page.tmpl" and call map["home.page.tmpl"]
	//
	// 139 - page-buildTemplateFromDisk.t:  about.page.tmpl
	// 174 - page-tpl.DefinedTemplates():  ; 
	//		defined templates are: "content", "base", "title.partial.tmpl", "about.page.tmpl", 
	//													 "css", "title", "footer", "base.layout.tmpl", 
	//													 "css.partial.tmpl", "footer.partial.tmpl"
	// So, all this is available when we render "about.page.tmpl" and call map["about.page.tmpl"]

	if ren.Debug {
		log.Println("Reading template", t, "from disk")
	}

	return tmpl, nil
}

// LoadLayoutsAndPartials accepts a slice of strings which should consist of the types of files
// that are either layouts or partials for templates.
// For example, if a layout file is named`base.layout.tmpl` and a partial is named `footer.partial.tmpl`,
// then we would pass:
//
//	[]string{".layout", ".partial"}
//
// Files anywhere in TemplateDir will be added the the Partials field of the Render type.
//
// Function returns:
//  [templates/base.layout.tmpl templates/css.partial.tmpl templates/footer.partial.tmpl]
func (ren *Render) LoadLayoutsAndPartials(fileTypes []string) error {
	fmt.Println("159 - page-LoadLayoutsAndPartials: ", fileTypes)
	// 159 - page-LoadLayoutsAndPartials:  [.layout .partial]
	var templates []string
	for _, t := range fileTypes {
		files, err := addTemplate(ren.TemplateDir, t)
		if err != nil {
			return err
		}
		templates = append(templates, files...)
	}
	ren.Partials = templates
	fmt.Println("171 - page-LoadLayoutsAndPartials: ", ren.Partials)
	// 171 - page-LoadLayoutsAndPartials:  [templates/base.layout.tmpl templates/css.partial.tmpl templates/footer.partial.tmpl]
	return nil
}

func addTemplate(path, fileType string) ([]string, error) {
	files, err := find(path, ".tmpl")
	if err != nil {
		return nil, err
	}
	var templates []string
	for _, x := range files {
		if strings.Contains(x, fileType) {
			templates = append(templates, x)
		}
	}
	return templates, nil
}

func find(root, ext string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			files = append(files, s)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
