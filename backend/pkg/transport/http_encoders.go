package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eaguilar88/deu/pkg/endorsements"
)

// Endorsements Encoders
func encodeGetEndorsementResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsements.GetEndorsementResponse)
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

func encodeGetEndorsementsResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsements.GetEndorsementsResponse)
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

func encodeCreateEndorsementResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsements.CreateEndorsementResponse)
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

func encodeUpdateEndorsementResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsements.UpdateEndorsementResponse)
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

func encodeDeleteEndorsementResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsements.DeleteEndorsementResponse)
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

// Endorsements Courses
func encodeGetCourseResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(endorsements.GetEndorsementResponse)
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
	resp, ok := untypedResp.(endorsements.GetEndorsementsResponse)
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
	resp, ok := untypedResp.(endorsements.CreateEndorsementResponse)
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
	resp, ok := untypedResp.(endorsements.UpdateEndorsementResponse)
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
	resp, ok := untypedResp.(endorsements.DeleteEndorsementResponse)
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
