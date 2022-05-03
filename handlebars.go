package defaults

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/infinytum/go-mojito"
	"github.com/infinytum/go-mojito/log"
	"github.com/infinytum/go-mojito/mojito/renderer"
	"github.com/infinytum/go-mojito/pkg/structures"
	"github.com/infinytum/raymond/v2"
)

var (
	raymondViewTemplateNotFound = "View Template '%s' not found or not readable"
)

func init() {
	raymond.ResolvePartial = resolvePartial
	raymond.RegisterHelper("extends", helperExtends)
	raymond.RegisterHelper("view", helperView)
	raymond.RegisterHelper("when", helperWhen)
	raymond.RegisterHelper("set", helperSet)
	raymond.RegisterHelper("formatdate", helperDate)
}

type handlebarsRenderer struct{}

// Render will load a template file and render the template
// within using the viewbag as a context
func (r *handlebarsRenderer) Render(view string, bag renderer.ViewBag) (string, error) {
	tpl, err := raymond.ParseFile(normalizeViewPath(view))
	if err != nil {
		log.Error(err)
		return fmt.Sprintf(raymondViewTemplateNotFound, view), err
	}
	return tpl.Exec(bag)
}

// newHandlebarsRenderer will return a new instance of the mojito handlebars renderer implementation
func newHandlebarsRenderer() mojito.Renderer {
	return &handlebarsRenderer{}
}

// helperExtends provides the template extension feature
func helperExtends(view string, options *raymond.Options) raymond.SafeString {
	tpl, err := raymond.ParseFile(normalizeViewPath(view))
	if err != nil {
		log.Error(err)
		return raymond.SafeString(fmt.Sprintf(raymondViewTemplateNotFound, view))
	}
	data, err := json.Marshal(options.Ctx())
	if err != nil {
		log.Error(err)
		return raymond.SafeString("Unable to encode context")
	}
	newBag := structures.NewMap[string, interface{}]()
	if err := json.Unmarshal(data, &newBag); err != nil {
		log.Error(err)
		return raymond.SafeString("Unable to decode context")
	}
	newBag.Set("subview", options.FnWith(options.Ctx()))
	return raymond.SafeString(tpl.MustExec(newBag))
}

// helperWhen provides a shorthand if inside expression blocks
func helperWhen(conditional bool, whenTrue interface{}, whenFalse interface{}, options *raymond.Options) interface{} {
	if conditional {
		return whenTrue
	}
	return whenFalse
}

// helperView provides dynamic view injection & rendering
func helperView(view string, bag interface{}) raymond.SafeString {
	tpl, err := raymond.ParseFile(normalizeViewPath(view))
	if err != nil {
		log.Error(err)
		return raymond.SafeString(fmt.Sprintf(raymondViewTemplateNotFound, view))
	}
	return raymond.SafeString(tpl.MustExec(bag))
}

func helperSet(propName string, propVal interface{}, options *raymond.Options) string {
	data, err := json.Marshal(options.Ctx())
	if err != nil {
		log.Error(err)
		return "Unable to encode context"
	}
	newBag := structures.NewMap[string, interface{}]()
	if err := json.Unmarshal(data, &newBag); err != nil {
		log.Error(err)
		return "Unable to decode context"
	}
	newBag.Set(propName, propVal)
	return options.FnWith(newBag)
}

func helperDate(date interface{}, format string) string {
	reflectType := reflect.TypeOf(date)
	if reflectType.Kind() == reflect.Int {
		date = int64(date.(int))
		reflectType = reflect.TypeOf(date)
	}
	if reflectType.Kind() == reflect.Int64 {
		stamp := date.(int64)
		return time.Unix(stamp, 0).Format(format)
	}

	if reflectType.AssignableTo(reflect.TypeOf(time.Time{})) {
		stamp := date.(time.Time)
		return stamp.Format(format)
	}
	return "Invalid date"
}

func resolvePartial(view string) *raymond.Partial {
	path := normalizeViewPath(view)
	tpl, err := raymond.ParseFile(path)
	if err != nil {
		log.Error(err)
		return nil
	}
	return raymond.NewPartial(view, path, tpl)
}

// normalizeViewPath ensures the path is within bounds and ends with .mojito
func normalizeViewPath(view string) string {
	path := mojito.ResourcesDir + TemplatePrefix + view + ".mojito"
	if strings.HasPrefix(path, mojito.ResourcesDir+TemplatePrefix) {
		return path
	}
	log.Warnf("Attempted path traversal to " + path)
	return mojito.ResourcesDir + TemplatePrefix
}
