// Code generated by ogen, DO NOT EDIT.

package scrapper

import (
	"net/http"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func encodeLinksDeleteResponse(response LinksDeleteRes, w http.ResponseWriter, span trace.Span) error {
	switch response := response.(type) {
	case *LinkResponse:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		span.SetStatus(codes.Ok, http.StatusText(200))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *LinksDeleteBadRequest:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(400)
		span.SetStatus(codes.Error, http.StatusText(400))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *LinksDeleteNotFound:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(404)
		span.SetStatus(codes.Error, http.StatusText(404))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *LinksDeleteTooManyRequests:
		w.WriteHeader(429)
		span.SetStatus(codes.Error, http.StatusText(429))

		return nil

	default:
		return errors.Errorf("unexpected response type: %T", response)
	}
}

func encodeLinksGetResponse(response LinksGetRes, w http.ResponseWriter, span trace.Span) error {
	switch response := response.(type) {
	case *ListLinksResponse:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		span.SetStatus(codes.Ok, http.StatusText(200))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *ApiErrorResponse:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(400)
		span.SetStatus(codes.Error, http.StatusText(400))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *LinksGetTooManyRequests:
		w.WriteHeader(429)
		span.SetStatus(codes.Error, http.StatusText(429))

		return nil

	default:
		return errors.Errorf("unexpected response type: %T", response)
	}
}

func encodeLinksPostResponse(response LinksPostRes, w http.ResponseWriter, span trace.Span) error {
	switch response := response.(type) {
	case *LinkResponse:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		span.SetStatus(codes.Ok, http.StatusText(200))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *ApiErrorResponse:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(400)
		span.SetStatus(codes.Error, http.StatusText(400))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *LinksPostTooManyRequests:
		w.WriteHeader(429)
		span.SetStatus(codes.Error, http.StatusText(429))

		return nil

	default:
		return errors.Errorf("unexpected response type: %T", response)
	}
}

func encodeTgChatIDDeleteResponse(response TgChatIDDeleteRes, w http.ResponseWriter, span trace.Span) error {
	switch response := response.(type) {
	case *TgChatIDDeleteOK:
		w.WriteHeader(200)
		span.SetStatus(codes.Ok, http.StatusText(200))

		return nil

	case *TgChatIDDeleteBadRequest:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(400)
		span.SetStatus(codes.Error, http.StatusText(400))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *TgChatIDDeleteNotFound:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(404)
		span.SetStatus(codes.Error, http.StatusText(404))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *TgChatIDDeleteTooManyRequests:
		w.WriteHeader(429)
		span.SetStatus(codes.Error, http.StatusText(429))

		return nil

	default:
		return errors.Errorf("unexpected response type: %T", response)
	}
}

func encodeTgChatIDPostResponse(response TgChatIDPostRes, w http.ResponseWriter, span trace.Span) error {
	switch response := response.(type) {
	case *TgChatIDPostOK:
		w.WriteHeader(200)
		span.SetStatus(codes.Ok, http.StatusText(200))

		return nil

	case *ApiErrorResponse:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(400)
		span.SetStatus(codes.Error, http.StatusText(400))

		e := new(jx.Encoder)
		response.Encode(e)
		if _, err := e.WriteTo(w); err != nil {
			return errors.Wrap(err, "write")
		}

		return nil

	case *TgChatIDPostTooManyRequests:
		w.WriteHeader(429)
		span.SetStatus(codes.Error, http.StatusText(429))

		return nil

	default:
		return errors.Errorf("unexpected response type: %T", response)
	}
}
