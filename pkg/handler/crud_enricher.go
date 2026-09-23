package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/mapper/entity_restapi/e2r"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/pkg/util"
	enricherapi "github.com/openflagr/flagr/swagger_gen/restapi/operations/enricher"
	"gorm.io/gorm"
)

// normalizeEnricherConfig rewrites a jev config's question names into the
// canonical `@jev_<name>` form before it is stored, so persisted property names
// do not depend on what the caller sent. Other namespaces pass through.
func normalizeEnricherConfig(namespace, configJSON string) (string, error) {
	if namespace != entity.EnricherNamespaceJev {
		return configJSON, nil
	}
	cfg, err := DecodeJevConfig(configJSON)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("encoding jev enricher config: %w", err)
	}
	return string(b), nil
}

// marshalEnricherConfig encodes a create/put request config into the stored
// JSON and validates the enricher definition (namespace + config schema).
func marshalEnricherConfig(flagID uint, namespace string, config any) (*entity.Enricher, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("encoding enricher config: %w", err)
	}
	normalized, err := normalizeEnricherConfig(namespace, string(b))
	if err != nil {
		return nil, err
	}
	e := &entity.Enricher{
		FlagID:     flagID,
		Namespace:  namespace,
		ConfigJSON: normalized,
	}
	if err := validateEnricher(e); err != nil {
		return nil, err
	}
	return e, nil
}

func findFlagEnricher(tx *gorm.DB, flagID uint, namespace string) (*entity.Enricher, error) {
	e := &entity.Enricher{}
	if err := tx.Where("flag_id = ? AND namespace = ?", flagID, namespace).First(e).Error; err != nil {
		return nil, err
	}
	return e, nil
}

func (c *crud) CreateEnricher(params enricherapi.CreateEnricherParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)

	e, err := marshalEnricherConfig(flagID, util.SafeString(params.Body.Namespace), params.Body.Config)
	if err != nil {
		return enricherapi.NewCreateEnricherDefault(400).WithPayload(ErrorMessage("%s", err))
	}

	var created *entity.Enricher
	err = commitFlagMutation(flagID, subject, notification.OperationCreate, notification.ComponentEnricher, func(tx *gorm.DB) (uint, mutationNotify, error) {
		if _, err := findFlagEnricher(tx, flagID, e.Namespace); err == nil {
			return 0, mutationNotify{}, fmt.Errorf("enricher namespace %q already exists on this flag", e.Namespace)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, mutationNotify{}, err
		}
		if err := tx.Create(e).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		created = e
		return flagID, mutationNotify{ComponentID: e.ID, ComponentKey: e.Namespace}, nil
	})
	if err != nil {
		return enricherapi.NewCreateEnricherDefault(400).WithPayload(ErrorMessage("%s", err))
	}

	resp := enricherapi.NewCreateEnricherOK()
	resp.SetPayload(e2r.MapEnricher(created))
	return resp
}

func (c *crud) PutEnricher(params enricherapi.PutEnricherParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)

	candidate, err := marshalEnricherConfig(flagID, params.Namespace, params.Body.Config)
	if err != nil {
		return enricherapi.NewPutEnricherDefault(400).WithPayload(ErrorMessage("%s", err))
	}

	var updated *entity.Enricher
	err = commitFlagMutation(flagID, subject, notification.OperationUpdate, notification.ComponentEnricher, func(tx *gorm.DB) (uint, mutationNotify, error) {
		e, err := findFlagEnricher(tx, flagID, candidate.Namespace)
		if err != nil {
			return 0, mutationNotify{}, err
		}
		e.ConfigJSON = candidate.ConfigJSON
		if err := tx.Save(e).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		updated = e
		return flagID, mutationNotify{ComponentID: e.ID, ComponentKey: e.Namespace}, nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return enricherapi.NewPutEnricherDefault(404).WithPayload(ErrorMessage("enricher %q not found", candidate.Namespace))
		}
		return enricherapi.NewPutEnricherDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	resp := enricherapi.NewPutEnricherOK()
	resp.SetPayload(e2r.MapEnricher(updated))
	return resp
}

func (c *crud) DeleteEnricher(params enricherapi.DeleteEnricherParams) middleware.Responder {
	flagID := util.SafeUint(params.FlagID)
	subject := getSubjectFromRequest(params.HTTPRequest)
	namespace := params.Namespace

	err := commitFlagMutation(flagID, subject, notification.OperationDelete, notification.ComponentEnricher, func(tx *gorm.DB) (uint, mutationNotify, error) {
		e, err := findFlagEnricher(tx, flagID, namespace)
		if err != nil {
			return 0, mutationNotify{}, err
		}
		if err := tx.Delete(e).Error; err != nil {
			return 0, mutationNotify{}, err
		}
		return flagID, mutationNotify{ComponentID: e.ID, ComponentKey: namespace}, nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return enricherapi.NewDeleteEnricherDefault(404).WithPayload(ErrorMessage("enricher %q not found", namespace))
		}
		return enricherapi.NewDeleteEnricherDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return enricherapi.NewDeleteEnricherOK()
}
