package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eaguilar88/deu/pkg/endorsments"
)

// Endorsments Encoders
func encodeGetEndorsmentResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.GetEndorsmentResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return nil
}

func encodeGetEndorsmentsResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.GetEndorsmentsResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return nil
}

func encodeCreateEndorsmentResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.CreateEndorsmentResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
	return nil
}

func encodeUpdateEndorsmentResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.UpdateEndorsmentResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusAccepted)
	w.Write(b)
	return nil
}

func encodeDeleteEndorsmentResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.DeleteEndorsmentResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusAccepted)
	w.Write(b)
	return nil
}

// Endorsments Courses
func encodeGetCourseResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.GetEndorsmentResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return nil
}

func encodeGetCoursesResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.GetEndorsmentsResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return nil
}

func encodeCreateCourseResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.CreateEndorsmentResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
	return nil
}

func encodeUpdateCourseResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.UpdateEndorsmentResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusAccepted)
	w.Write(b)
	return nil
}

func encodeDeleteCourseResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsments.DeleteEndorsmentResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusAccepted)
	w.Write(b)
	return nil
}
