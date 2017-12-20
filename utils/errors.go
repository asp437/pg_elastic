package utils

import (
	"fmt"
)

type ElasticError interface {
	error
	Type() string
	Reason() string
	FormatErrorResponse() interface{}
}

type ElasticErrorGeneral struct {
	TypeVal   string `json:"type"`
	ReasonVal string `json:"reason"`
}

type ElasticErrorGeneralResponse struct {
	RootCause []ElasticErrorGeneral `json:"root_cause"`
	ElasticErrorGeneral
}

type ElasticErrorBulk struct {
	ElasticErrorGeneral
	Index     string `json:"index"`
	Shard     string `json:"shard"`
	IndexUUID string `json:"index_uuid"`
}

type JSONWrongFormatError struct {
	ElasticErrorGeneral
}

type DBQueryError struct {
	ElasticErrorGeneral
}

type InternalIOError struct {
	ElasticErrorGeneral
}

type InternalError struct {
	ElasticErrorGeneral
}

type IllegalQueryError struct {
	ElasticErrorGeneral
}

func (err *ElasticErrorGeneral) Error() string {
	return fmt.Sprintf("Error type: %s, Reason: %s", err.Type(), err.Reason())
}

func (err *ElasticErrorGeneral) Type() string {
	return err.TypeVal
}

func (err *ElasticErrorGeneral) Reason() string {
	return err.ReasonVal
}

func (err *ElasticErrorGeneral) FormatErrorResponse() interface{} {
	output := make(map[string]interface{})
	errorDesc := ElasticErrorGeneralResponse{}
	errorDesc.RootCause = []ElasticErrorGeneral{ElasticErrorGeneral{err.Type(), err.Reason()}}
	errorDesc.ReasonVal = err.Reason()
	errorDesc.TypeVal = err.Type()
	output["error"] = errorDesc
	output["status"] = 500
	return output
}

func (err *ElasticErrorBulk) FormatErrorResponse() interface{} {
	output := make(map[string]interface{})
	output["error"] = err
	return output
}

func NewJSONWrongFormatError(reason string) *JSONWrongFormatError {
	return &JSONWrongFormatError{ElasticErrorGeneral{"json_parse_exception", reason}}
}

func NewDBQueryError(reason string) *DBQueryError {
	return &DBQueryError{ElasticErrorGeneral{"db_query_exception", reason}}
}

func NewInternalIOError(reason string) *InternalIOError {
	return &InternalIOError{ElasticErrorGeneral{"internal_io_exception", reason}}
}

func NewInternalError(reason string) *InternalError {
	return &InternalError{ElasticErrorGeneral{"internal_exception", reason}}
}

func NewIllegalQueryError(reason string) *IllegalQueryError {
	return &IllegalQueryError{ElasticErrorGeneral{"illegal_query_exception", reason}}
}

func NewElasticErrorBulk(err ElasticError, index, shard, indexUUID string) *ElasticErrorBulk {
	output := &ElasticErrorBulk{}
	output.Index = index
	output.IndexUUID = indexUUID
	output.Shard = shard
	output.ReasonVal = err.Reason()
	output.TypeVal = err.Type()
	return output
}
