package utils

import (
	"fmt"
)

// ElasticError is a basic interface for any kind of errors produced by pg-elastic and can be represented as a JSON error report
type ElasticError interface {
	error
	Type() string
	Reason() string
	FormatErrorResponse() interface{}
}

// ElasticErrorGeneral represents a response format for JSON error report 'cause' part
type ElasticErrorGeneral struct {
	TypeVal   string `json:"type"`
	ReasonVal string `json:"reason"`
}

// ElasticErrorGeneralResponse represents a response format for JSON error report
type ElasticErrorGeneralResponse struct {
	RootCause []ElasticErrorGeneral `json:"root_cause"`
	ElasticErrorGeneral
}

// ElasticErrorBulk represents a response format for JSON error report for bulk requests
type ElasticErrorBulk struct {
	ElasticErrorGeneral
	Index     string `json:"index"`
	Shard     string `json:"shard"`
	IndexUUID string `json:"index_uuid"`
}

// JSONWrongFormatError is error caused by illegal JSON input
type JSONWrongFormatError struct {
	ElasticErrorGeneral
}

// DBQueryError is error caused by any internal error of the database
type DBQueryError struct {
	ElasticErrorGeneral
}

// InternalIOError is error caused by any other IO operations
type InternalIOError struct {
	ElasticErrorGeneral
}

// InternalError is error caused by pg-elastic itself
type InternalError struct {
	ElasticErrorGeneral
}

// IllegalQueryError is error caused by misconstructed query
type IllegalQueryError struct {
	ElasticErrorGeneral
}

func (err *ElasticErrorGeneral) Error() string {
	return fmt.Sprintf("Error type: %s, Reason: %s", err.Type(), err.Reason())
}

// Type returns name of the error type
func (err *ElasticErrorGeneral) Type() string {
	return err.TypeVal
}

// Reason returns the error reason
func (err *ElasticErrorGeneral) Reason() string {
	return err.ReasonVal
}

// FormatErrorResponse generates an JSON output for an error
func (err *ElasticErrorGeneral) FormatErrorResponse() interface{} {
	output := make(map[string]interface{})
	errorDesc := ElasticErrorGeneralResponse{}
	errorDesc.RootCause = []ElasticErrorGeneral{{err.Type(), err.Reason()}}
	errorDesc.ReasonVal = err.Reason()
	errorDesc.TypeVal = err.Type()
	output["error"] = errorDesc
	output["status"] = 500
	return output
}

// FormatErrorResponse generates an JSON output for an error
func (err *ElasticErrorBulk) FormatErrorResponse() interface{} {
	output := make(map[string]interface{})
	output["error"] = err
	return output
}

// NewJSONWrongFormatError creates a new instance of JSONWrongFormatError
func NewJSONWrongFormatError(reason string) *JSONWrongFormatError {
	return &JSONWrongFormatError{ElasticErrorGeneral{"json_parse_exception", reason}}
}

// NewDBQueryError creates a new instance of DBQueryError
func NewDBQueryError(reason string) *DBQueryError {
	return &DBQueryError{ElasticErrorGeneral{"db_query_exception", reason}}
}

// NewInternalIOError creates a new instance of InternalIOError
func NewInternalIOError(reason string) *InternalIOError {
	return &InternalIOError{ElasticErrorGeneral{"internal_io_exception", reason}}
}

// NewInternalError creates a new instance of InternalError
func NewInternalError(reason string) *InternalError {
	return &InternalError{ElasticErrorGeneral{"internal_exception", reason}}
}

// NewIllegalQueryError creates a new instance of IllegalQueryError
func NewIllegalQueryError(reason string) *IllegalQueryError {
	return &IllegalQueryError{ElasticErrorGeneral{"illegal_query_exception", reason}}
}

// NewElasticErrorBulk creates a new instance of ElasticErrorBulk
func NewElasticErrorBulk(err ElasticError, index, shard, indexUUID string) *ElasticErrorBulk {
	output := &ElasticErrorBulk{}
	output.Index = index
	output.IndexUUID = indexUUID
	output.Shard = shard
	output.ReasonVal = err.Reason()
	output.TypeVal = err.Type()
	return output
}
