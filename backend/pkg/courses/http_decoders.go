package courses

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eaguilar88/deu/pkg/transport"
	"github.com/gorilla/mux"
)

func decodeGetCourseRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	courseID, ok := vars[transport.ParamCourseID]
	if !ok {
		return nil, fmt.Errorf("missing required param: %s", transport.ParamCourseID)
	}

	return GetCourseRequest{ID: courseID}, nil
}

func decodeGetCoursesRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	queryScope, err := transport.NewQueryScopeFromURL(r.URL)
	if err != nil {
		return nil, fmt.Errorf("error decoding query string: %w", err) // Wrap error
	}

	return GetCoursesRequest{PageScope: queryScope}, nil
}

func decodeCreateCourseRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	var req CreateCourseRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, fmt.Errorf("error decoding request body: %w", err) // Wrap error
	}
	defer r.Body.Close() // Close request body

	return req, nil
}

func decodeUpdateCourseRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	courseIDStr, ok := vars[transport.ParamCourseID]
	if !ok {
		return nil, fmt.Errorf("missing required param: %s", transport.ParamCourseID)
	}

	_, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid course ID: %w", err) // Wrap error
	}

	var req UpdateCourseRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, fmt.Errorf("error decoding request body: %w", err) // Wrap error
	}
	defer r.Body.Close() // Close request body
	req.ID = courseIDStr //or courseID

	return req, nil
}

func decodeDeleteCourseRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	courseID, ok := vars[transport.ParamCourseID]
	if !ok {
		return nil, fmt.Errorf("missing required param: %s", transport.ParamCourseID)
	}
	return DeleteCourseRequest{ID: courseID}, nil
}
