package handler

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jinzhu/gorm"
)

func (c *crud) CreateFlag(params flag.CreateFlagParams) middleware.Responder {
	f := &entity.Flag{}
	if params.Body != nil {
		f.Description = util.SafeString(params.Body.Description)
		f.CreatedBy = getSubjectFromRequest(params.HTTPRequest)

		key, err := entity.CreateFlagKey(params.Body.Key)
		if err != nil {
			return flag.NewCreateFlagDefault(400).WithPayload(
				ErrorMessage("cannot create flag. %s", err))
		}
		f.Key = key
	}

	tx := getDB().Begin()

	if err := tx.Create(f).Error; err != nil {
		tx.Rollback()
		return flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage("cannot create flag. %s", err))
	}

	// if params.Body.Template == "simple_boolean_flag" {
	// 	if err := LoadSimpleBooleanFlagTemplate(f, tx); err != nil {
	// 		tx.Rollback()
	// 		return flag.NewCreateFlagDefault(500).WithPayload(
	// 			ErrorMessage("cannot create flag. %s", err))
	// 	}
	// } else
	if params.Body.Template != "" {
		found := false
		for name, templateStr := range config.Config.CustomNewFlagTemplates {
			found = name == params.Body.Template
			if found {
				tmpl, err := template.New("test").Parse(templateStr)
				if err != nil {
					panic(err)
				}

				descriptionJsonParams := map[string]interface{}{}
				if err := json.Unmarshal([]byte(*params.Body.Description), &descriptionJsonParams); err != nil {
					panic(err)
				}

				var tpl bytes.Buffer
				err = tmpl.Execute(&tpl, descriptionJsonParams)
				if err != nil {
					panic(err)
				}

				flagJsonFromTemplate := tpl.String()
				// fmt.Printf("++++++++++++Before %v\n++++++++++\n", f)
				if err := json.Unmarshal([]byte(flagJsonFromTemplate), &f); err != nil {
					panic(err)
				}
				// fmt.Printf("++++++++++++After %v\n++++++++++\n", f)

				if err := LoadDependentObjectsFromFlag(f, tx); err != nil {
					tx.Rollback()
					panic(err)
				}

				break
			}
		}
		if !found {
			return flag.NewCreateFlagDefault(400).WithPayload(
				ErrorMessage("unknown value for template: %s", params.Body.Template))
		}
	}

	err := tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return flag.NewCreateFlagDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := flag.NewCreateFlagOK()
	payload, err := e2rMapFlag(f)
	if err != nil {
		return flag.NewCreateFlagDefault(500).WithPayload(
			ErrorMessage("cannot map flag. %s", err))
	}
	resp.SetPayload(payload)

	entity.SaveFlagSnapshot(getDB(), f.ID, getSubjectFromRequest(params.HTTPRequest))

	return resp
}

func LoadDependentObjectsFromFlag(flag *entity.Flag, tx *gorm.DB) error {
	for _, tag := range flag.Tags {
		t := &entity.Tag{}
		t.Flags = []*entity.Flag{&entity.Flag{Model: gorm.Model{ID: flag.ID}}}
		// fmt.Printf("We're creating! \n %v +++++%v++++++\n", flag, tag.Value)
		t.Value = tag.Value

		// fmt.Printf("-----------%v---------\n", t)
		if err := tx.Create(t).Error; err != nil {
			return err
		}
	}
	return nil
}

// LoadSimpleBooleanFlagTemplate loads the simple boolean flag template into
// a new flag. It creates a single segment, variant ('on'), and distribution.
func LoadSimpleBooleanFlagTemplate(flag *entity.Flag, tx *gorm.DB) error {
	// Create our default segment
	s := &entity.Segment{}
	s.FlagID = flag.ID
	s.RolloutPercent = uint(100)
	s.Rank = entity.SegmentDefaultRank

	if err := tx.Create(s).Error; err != nil {
		return err
	}

	// .. and our default Variant
	v := &entity.Variant{}
	v.FlagID = flag.ID
	v.Key = "on"

	if err := tx.Create(v).Error; err != nil {
		return err
	}

	// .. and our default Distribution
	d := &entity.Distribution{}
	d.SegmentID = s.ID
	d.VariantID = v.ID
	d.VariantKey = v.Key
	d.Percent = uint(100)

	if err := tx.Create(d).Error; err != nil {
		return err
	}

	s.Distributions = append(s.Distributions, *d)
	flag.Variants = append(flag.Variants, *v)
	flag.Segments = append(flag.Segments, *s)

	return nil
}
