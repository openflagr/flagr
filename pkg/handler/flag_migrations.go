package handler

import (
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/mapper/entity_restapi/r2e"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gorm.io/gorm/clause"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
)

type FlagMigrationOptions struct {
	Run  bool   `long:"migrations" short:"m" description:"Run Flag Migrations and Exit"`
	Path string `long:"migrationPath" description:"Migration files path" env:"FLAGR_MIGRATION_PATH"`
}

func FlagMigrations(path string) error {
	return migrateFromDir(path)
}

func migrateFromDir(dir string) error {
	tx := getDB()
	fms := []entity.FlagMigration{}

	files, err := os.ReadDir(dir)
	if err != nil {
		logrus.WithField("err", err).Errorf("cannot read directory for migrations: %v", err)
		return err
	}

	completedFiles := make(map[string]bool)

	tx.Find(&fms)

	for _, fm := range fms {
		completedFiles[fm.Name] = true
	}

	completed := 0
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".yaml") && !completedFiles[filename] {
			flags, err := readMigrationFile(dir, filename)
			if err != nil {
				continue
			}

			completed = migrateFlags(flags, filename) + completed
		}

	}

	logrus.Infof("%d new migrations completed", completed)
	return nil
}

func migrateFlags(flags []models.Flag, filename string) int {
	for _, flagModel := range flags {
		if f, err := migrateFlag(flagModel); err == nil {
			entity.SaveFlagSnapshot(getDB(), util.SafeUint(f.ID), "migration "+filename)
		}
	}

	fm := entity.FlagMigration{Name: filename}
	getDB().Create(&fm)
	return 1
}

func migrateFlag(flagModel models.Flag) (*entity.Flag, error) {
	flag := &entity.Flag{
		EntityType:         flagModel.EntityType,
		Key:                flagModel.Key,
		Description:        util.SafeString(flagModel.Description),
		DataRecordsEnabled: cast.ToBool(flagModel.DataRecordsEnabled),
		Enabled:            cast.ToBool(flagModel.Enabled),
		Notes:              util.SafeStringWithDefault(flagModel.Notes, ""),
	}

	tx := getDB()

	//upsert
	tx.Where(entity.Flag{Key: flagModel.Key}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		UpdateAll: true,
	}).Create(&flag)

	tx = entity.PreloadSegmentsVariantsTags(getDB())

	tx.Where(entity.Flag{Key: flagModel.Key}).First(&flag)

	// delete the association
	deleteTags(flag)

	// delete the entity (keeps associations
	deleteVariants(flag)
	deleteSegments(flag)

	addTags(flagModel, flag)

	// used to map variant to distribution
	variantMap := saveVariants(flagModel, flag)

	saveSegments(flagModel, flag, variantMap)

	return flag, nil
}

func saveSegments(flagModel models.Flag, flag *entity.Flag, variantMap map[string]int64) {
	for _, segmentModel := range flagModel.Segments {
		segment := entity.Segment{
			Description:    util.SafeString(segmentModel.Description),
			RolloutPercent: uint(*segmentModel.RolloutPercent),
			Rank:           uint(*segmentModel.Rank),
		}
		// save segment to flag
		getDB().Model(flag).Association("Segments").Append(&segment)
		// map distribution to variant
		for _, distribution := range segmentModel.Distributions {
			variantId := variantMap[util.SafeString(distribution.VariantKey)]
			distribution.VariantID = &variantId
		}
		segment.Distributions = r2e.MapDistributions(segmentModel.Distributions, segment.ID)
		segment.Constraints = r2e.MapConstraints(segmentModel.Constraints, segment.ID)
		getDB().Save(&segment)
	}
}

func saveVariants(flagModel models.Flag, flag *entity.Flag) map[string]int64 {
	variantMap := make(map[string]int64)

	// add variants
	for _, variantModel := range flagModel.Variants {
		a, _ := r2e.MapAttachment(variantModel.Attachment)
		variant := entity.Variant{
			Key:        util.SafeString(variantModel.Key),
			Attachment: a,
		}
		getDB().Model(flag).Association("Variants").Append(&variant)
		variantMap[util.SafeString(variant.Key)] = int64(variant.ID)
	}
	return variantMap
}

func addTags(flagModel models.Flag, flag *entity.Flag) {
	// add tags
	for _, tagModel := range flagModel.Tags {
		t := &entity.Tag{}
		t.Value = util.SafeString(tagModel.Value)
		getDB().Where("value = ?", util.SafeString(tagModel.Value)).Find(t)
		getDB().Model(flag).Association("Tags").Append(t)
	}
}

func deleteSegments(flag *entity.Flag) {
	for _, segmentsModel := range flag.Segments {
		getDB().Select("Constraints", "Distributions").Delete(&entity.Segment{}, segmentsModel.ID)
	}
}

func deleteVariants(flag *entity.Flag) {
	for _, variantsModel := range flag.Variants {
		v := &entity.Variant{}
		v.ID = variantsModel.ID
		getDB().Delete(&entity.Variant{}, variantsModel.ID)
	}
}

func deleteTags(flag *entity.Flag) {
	for _, tagsModel := range flag.Tags {
		t := &entity.Tag{}
		t.ID = uint(tagsModel.ID)
		getDB().Model(flag).Association("Tags").Delete(t)
	}
}

func readMigrationFile(dir string, fileName string) ([]models.Flag, error) {
	f := dir + string(os.PathSeparator) + fileName
	contents, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	var flags []models.Flag
	err = yaml.Unmarshal([]byte(contents), &flags)
	if err != nil {
		return nil, err
	}
	return flags, nil
}
