package courses

import (
	"github.com/go-kit/kit/endpoint"
	kitHTTP "github.com/go-kit/kit/transport/http"
)

// Course Endpoints
func GetCourseHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeGetCourseRequestHTTP,
		encodeGetCourseResponseHTTP,
		options...,
	)
}

func GetCoursesHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeGetCoursesRequestHTTP,
		encodeGetCoursesResponseHTTP,
		options...,
	)
}

func CreateCourseHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeCreateCourseRequestHTTP,
		encodeCreateCourseResponseHTTP,
		options...,
	)
}

func UpdateCourseHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeUpdateCourseRequestHTTP,
		encodeUpdateCourseResponseHTTP,
		options...,
	)
}

func DeleteCourseHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeDeleteCourseRequestHTTP,
		encodeDeleteCourseResponseHTTP,
		options...,
	)
}
