// Code generated by ogen, DO NOT EDIT.

package scrapper

import (
	"net/http"
	"net/url"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/conv"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/uri"
	"github.com/ogen-go/ogen/validate"
)

// LinksDeleteParams is parameters of DELETE /links operation.
type LinksDeleteParams struct {
	TgChatID int64
}

func unpackLinksDeleteParams(packed middleware.Parameters) (params LinksDeleteParams) {
	{
		key := middleware.ParameterKey{
			Name: "Tg-Chat-Id",
			In:   "header",
		}
		params.TgChatID = packed[key].(int64)
	}
	return params
}

func decodeLinksDeleteParams(args [0]string, argsEscaped bool, r *http.Request) (params LinksDeleteParams, _ error) {
	h := uri.NewHeaderDecoder(r.Header)
	// Decode header: Tg-Chat-Id.
	if err := func() error {
		cfg := uri.HeaderParameterDecodingConfig{
			Name:    "Tg-Chat-Id",
			Explode: false,
		}
		if err := h.HasParam(cfg); err == nil {
			if err := h.DecodeParam(cfg, func(d uri.Decoder) error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToInt64(val)
				if err != nil {
					return err
				}

				params.TgChatID = c
				return nil
			}); err != nil {
				return err
			}
		} else {
			return err
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "Tg-Chat-Id",
			In:   "header",
			Err:  err,
		}
	}
	return params, nil
}

// LinksGetParams is parameters of GET /links operation.
type LinksGetParams struct {
	TgChatID int64
	Tag      OptString
}

func unpackLinksGetParams(packed middleware.Parameters) (params LinksGetParams) {
	{
		key := middleware.ParameterKey{
			Name: "Tg-Chat-Id",
			In:   "header",
		}
		params.TgChatID = packed[key].(int64)
	}
	{
		key := middleware.ParameterKey{
			Name: "tag",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.Tag = v.(OptString)
		}
	}
	return params
}

func decodeLinksGetParams(args [0]string, argsEscaped bool, r *http.Request) (params LinksGetParams, _ error) {
	q := uri.NewQueryDecoder(r.URL.Query())
	h := uri.NewHeaderDecoder(r.Header)
	// Decode header: Tg-Chat-Id.
	if err := func() error {
		cfg := uri.HeaderParameterDecodingConfig{
			Name:    "Tg-Chat-Id",
			Explode: false,
		}
		if err := h.HasParam(cfg); err == nil {
			if err := h.DecodeParam(cfg, func(d uri.Decoder) error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToInt64(val)
				if err != nil {
					return err
				}

				params.TgChatID = c
				return nil
			}); err != nil {
				return err
			}
		} else {
			return err
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "Tg-Chat-Id",
			In:   "header",
			Err:  err,
		}
	}
	// Decode query: tag.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "tag",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotTagVal string
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToString(val)
					if err != nil {
						return err
					}

					paramsDotTagVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.Tag.SetTo(paramsDotTagVal)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "tag",
			In:   "query",
			Err:  err,
		}
	}
	return params, nil
}

// LinksPostParams is parameters of POST /links operation.
type LinksPostParams struct {
	TgChatID int64
}

func unpackLinksPostParams(packed middleware.Parameters) (params LinksPostParams) {
	{
		key := middleware.ParameterKey{
			Name: "Tg-Chat-Id",
			In:   "header",
		}
		params.TgChatID = packed[key].(int64)
	}
	return params
}

func decodeLinksPostParams(args [0]string, argsEscaped bool, r *http.Request) (params LinksPostParams, _ error) {
	h := uri.NewHeaderDecoder(r.Header)
	// Decode header: Tg-Chat-Id.
	if err := func() error {
		cfg := uri.HeaderParameterDecodingConfig{
			Name:    "Tg-Chat-Id",
			Explode: false,
		}
		if err := h.HasParam(cfg); err == nil {
			if err := h.DecodeParam(cfg, func(d uri.Decoder) error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToInt64(val)
				if err != nil {
					return err
				}

				params.TgChatID = c
				return nil
			}); err != nil {
				return err
			}
		} else {
			return err
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "Tg-Chat-Id",
			In:   "header",
			Err:  err,
		}
	}
	return params, nil
}

// TgChatIDDeleteParams is parameters of DELETE /tg-chat/{id} operation.
type TgChatIDDeleteParams struct {
	ID int64
}

func unpackTgChatIDDeleteParams(packed middleware.Parameters) (params TgChatIDDeleteParams) {
	{
		key := middleware.ParameterKey{
			Name: "id",
			In:   "path",
		}
		params.ID = packed[key].(int64)
	}
	return params
}

func decodeTgChatIDDeleteParams(args [1]string, argsEscaped bool, r *http.Request) (params TgChatIDDeleteParams, _ error) {
	// Decode path: id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToInt64(val)
				if err != nil {
					return err
				}

				params.ID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// TgChatIDPostParams is parameters of POST /tg-chat/{id} operation.
type TgChatIDPostParams struct {
	ID int64
}

func unpackTgChatIDPostParams(packed middleware.Parameters) (params TgChatIDPostParams) {
	{
		key := middleware.ParameterKey{
			Name: "id",
			In:   "path",
		}
		params.ID = packed[key].(int64)
	}
	return params
}

func decodeTgChatIDPostParams(args [1]string, argsEscaped bool, r *http.Request) (params TgChatIDPostParams, _ error) {
	// Decode path: id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToInt64(val)
				if err != nil {
					return err
				}

				params.ID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}
