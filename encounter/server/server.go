package server

import (
	"context"
	"github.com/gabrielseibel1/gaef/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	encounterCreator      EncounterCreator
	encounterReaderByUser EncounterReaderByUser
	encounterReaderByID   EncounterReaderByID
	encounterUpdater      EncounterUpdater
	encounterDeleter      EncounterDeleter
	encounterConfirmer    EncounterConfirmer
}

func New(
	encounterCreator EncounterCreator,
	encounterReaderByUser EncounterReaderByUser,
	encounterReaderByID EncounterReaderByID,
	encounterUpdater EncounterUpdater,
	encounterDeleter EncounterDeleter,
	encounterConfirmer EncounterConfirmer,
) Server {
	return Server{
		encounterCreator:      encounterCreator,
		encounterReaderByUser: encounterReaderByUser,
		encounterReaderByID:   encounterReaderByID,
		encounterUpdater:      encounterUpdater,
		encounterDeleter:      encounterDeleter,
		encounterConfirmer:    encounterConfirmer,
	}
}

func (s Server) CreateEncounterHandler() gin.HandlerFunc {
	return jsonHandler(func(c *gin.Context) Result {
		e, err := encounter(c)
		if err != nil {
			return errorResult{s: http.StatusBadRequest, e: err}
		}

		t := token(c)
		return s.encounterCreator.CreateEncounter(c, t, e)
	})
}

func (s Server) ReadUserEncountersHandler() gin.HandlerFunc {
	return jsonHandler(func(c *gin.Context) Result {
		id := userID(c)
		return s.encounterReaderByUser.ReadUserEncounters(c, id)
	})
}

func (s Server) ReadEncounterHandler() gin.HandlerFunc {
	return jsonHandler(func(c *gin.Context) Result {
		uID, eID := userID(c), encID(c)
		return s.encounterReaderByID.ReadEncounter(c, uID, eID)
	})
}

func (s Server) UpdateEncounterHandler() gin.HandlerFunc {
	return jsonHandler(func(c *gin.Context) Result {
		e, err := encounter(c)
		if err != nil {
			return errorResult{s: http.StatusBadRequest, e: err}
		}

		uID, eID := userID(c), encID(c)
		return s.encounterUpdater.UpdateEncounter(c, uID, eID, e)
	})
}

func (s Server) DeleteEncounterHandler() gin.HandlerFunc {
	return jsonHandler(func(c *gin.Context) Result {
		uID, eID := userID(c), encID(c)
		return s.encounterDeleter.DeleteEncounter(c, uID, eID)
	})
}

func (s Server) ConfirmEncounterHandler() gin.HandlerFunc {
	return jsonHandler(func(c *gin.Context) Result {
		uID, eID := userID(c), encID(c)
		return s.encounterConfirmer.ConfirmEncounter(c, uID, eID)
	})
}

func jsonHandler(getResult func(c *gin.Context) Result) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := getResult(c)
		c.JSON(result.S(), gin.H{result.K(): result.V()})
	}
}

func encounter(ctx *gin.Context) (types.Encounter, error) {
	var e types.Encounter
	if err := ctx.ShouldBindJSON(&e); err != nil {
		return types.Encounter{}, err
	}
	return e, nil
}

func token(ctx *gin.Context) string {
	return ctx.GetString("token")
}

func userID(ctx *gin.Context) string {
	return ctx.GetString("userID")
}

func encID(ctx *gin.Context) string {
	return ctx.Param(EncIDParam)
}

var EncIDParam = "encounter-id"

type Result interface {
	S() int
	K() string
	V() any
}

type EncounterCreator interface {
	CreateEncounter(ctx context.Context, token string, e types.Encounter) Result
}

type EncounterReaderByUser interface {
	ReadUserEncounters(ctx context.Context, userID string) Result
}

type EncounterReaderByID interface {
	ReadEncounter(ctx context.Context, userID, encID string) Result
}

type EncounterUpdater interface {
	UpdateEncounter(ctx context.Context, userID string, encID string, e types.Encounter) Result
}

type EncounterDeleter interface {
	DeleteEncounter(ctx context.Context, userID, encID string) Result
}

type EncounterConfirmer interface {
	ConfirmEncounter(ctx context.Context, userID string, encID string) Result
}

type errorResult struct {
	s int
	e error
}

func (r errorResult) K() string {
	return "error"
}

func (r errorResult) V() any {
	return r.e.Error()
}

func (r errorResult) S() int {
	return r.s
}
